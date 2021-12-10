package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

// ConfigCmd can configure a fabric per a pre-defined configuration file.
type ConfigCmd struct {
	Path   string `kong:"arg,type='existingFile',help='Path to configuration file'"`
	DryRun bool   `kong:"optional,help='Perform a dry run of the configuration'"`
}

// The below structures are used to decode the configuration file. Be careful of making
// changes as these must also be reflected in the static config files.

// Domain referes to the Host Virtualization Domain.
type Domain struct {
	// Name identifies the Domain name. Domain names can be shared across different
	// devices or unique to a single Device.
	Name string

	// Port is the Upstream Port (USP) of the Domain. End-points are bound to
	// this port.
	Port uint8
}

// Device referes to a Switchtec device. It can have a series of Domains and
// any number of End-Points.
type Device struct {
	ID       int32
	Metadata struct {
		Name string
	}
	Domains   []Domain
	Endpoints []uint8 `yaml:",flow"`
}

// Option can be any key:value string pair applied to the configuration file.
type Option map[string]string

// ConfigFile is the top-level structure
type ConfigFile struct {
	Version  string
	Metadata struct {
		Name string
	}
	Options []Option `yaml:",omitempty"`
	Devices []Device
}

type configOptions struct {
	fabricCtrlMode fabricControllerMode
	virtMgmtMode   virtualizationManagmentMode
}

// Run will run the Config command.
func (cmd *ConfigCmd) Run() error {
	configData, err := ioutil.ReadFile(cmd.Path)
	if err != nil {
		return err
	}

	conf := ConfigFile{}

	fmt.Printf("Running Config %s...\n", path.Base(cmd.Path))
	if err := yaml.Unmarshal(configData, &conf); err != nil {
		return err
	}

	fmt.Printf("Config Loaded.\n")
	fmt.Printf("  Version: %s\n", conf.Version)
	fmt.Printf("  Name: %s\n", conf.Metadata.Name)

	fmt.Printf("Validate Config...\n")
	if err := validateConfig(conf); err != nil {
		return err
	}

	fmt.Print("Loading Options...\n")
	opts, err := loadOptions(conf)
	if err != nil {
		return err
	}

	totalDomains := 1 // Start with one domain being the Rabbit
	for _, d := range conf.Devices {
		totalDomains += len(d.Domains) - 1 // Sum the remaining domains (minus the Rabbit)
	}

	for deviceIndex, device := range conf.Devices {

		fmt.Printf("Locate Device Index: %d ID: %d...\n", deviceIndex, device.ID)
		devPath, err := locateDevice(device.ID, opts.fabricCtrlMode)
		if err != nil {
			return err
		}

		fmt.Printf("Device %d Found %s\n", device.ID, path.Base(devPath))
		dev, err := switchtec.Open(devPath)
		if err != nil {
			return err
		}
		defer dev.Close()

		fmt.Printf("Opening Virtualization Managment Controller...\n")
		virtMgmtCtrl, err := openVirtualizationManagementController(dev, opts.virtMgmtMode)
		if err != nil {
			return err
		}

		fmt.Printf("Generating Device Binding Table...\n")
		type Binding struct {
			hostSwIndex uint8
			domain      Domain
		}

		bindingTable := make([]Binding, totalDomains)

		// First half are the domains local to this device
		for i, domain := range device.Domains {
			bindingTable[i].hostSwIndex = uint8(device.ID)
			bindingTable[i].domain = domain
		}

		// Remainder are all the domains reached through the fabric
		index := len(device.Domains)
		for _, d := range conf.Devices {
			if d.ID == device.ID {
				continue
			}

			for i, domain := range d.Domains {
				if i == 0 { // Skip the interconnect
					continue
				}
				bindingTable[index].hostSwIndex = uint8(d.ID)
				bindingTable[index].domain = domain

				index++
			}
		}

		fmt.Printf("Enumerating End-Point Devices...\n")

		for epIndex, epPort := range device.Endpoints {
			fmt.Printf("Processing End-Point %d: Port %d\n", epIndex, epPort)

			processEndpoint := func(epIndex int, bindings []Binding, dryRun bool) func(*switchtec.DumpEpPortDevice) error {

				return func(epPort *switchtec.DumpEpPortDevice) error {

					if switchtec.EpPortType(epPort.Hdr.Typ) != switchtec.DeviceEpPortType {
						fmt.Printf("Warning: No device attached to end-point.\n")
						return nil
					}

					pfpdfid := epPort.Ep.Functions[0].PDFID

					for _, function := range epPort.Ep.Functions {

						if function.SRIOVCapPF != 0 {
							fmt.Printf("Physical Function. Skipping\n")
							continue
						}

						if function.FunctionID > uint16(len(bindings)) {
							fmt.Printf("No Domain Available.\n")
							continue
						}

						if err := configureVirtualizationManagementController(dev, virtMgmtCtrl, pfpdfid, function.FunctionID, dryRun); err != nil {
							return err
						}

						if function.Bound != 0 {
							fmt.Printf("Alread Bound: PAX: %d PhyPort: %d LogPort %d\n", function.BoundPAXID, function.BoundHVDPhyPID, function.BoundHVDLogPID)
							continue
						}

						binding := bindings[function.FunctionID-1]
						hostSwIndex := binding.hostSwIndex
						hostPhysPortID := binding.domain.Port
						hostLogPortID := uint8(epIndex)
						pdfid := function.PDFID
						fmt.Printf("Performing bind to %s: idx: %d phyPort %d logPort %d PDFID %#04x...", binding.domain.Name, hostSwIndex, hostPhysPortID, hostLogPortID, pdfid)

						if !dryRun {
							if err := dev.Bind(hostSwIndex, hostPhysPortID, hostLogPortID, pdfid); err != nil {
								fmt.Printf(" Error: %s\n", err)
								return err
							}
						}

						fmt.Printf(" Complete.\n")

					}
					return nil
				}
			}

			err := dev.GfmsEpPortDeviceEnumerate(uint8(epPort), processEndpoint(epIndex, bindingTable, cmd.DryRun))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateConfig(conf ConfigFile) error {

	configError := func(format string, a ...interface{}) error {
		return fmt.Errorf("Configuration Error: "+format, a...)
	}

	if conf.Version != "v1" {
		return configError("Unsupported version '%s'", conf.Version)
	}

	devs := len(conf.Devices)
	if devs != 2 {
		return configError("Unsupported device count. Expected: 2 Actual %d", devs)
	}

	for _, device := range conf.Devices {
		if len(device.Domains) == 0 {
			return configError("Atleast one domain needed for device %s", device.Metadata.Name)
		}
	}

	// Domain zero is special and is expected to map to the same Rabbit node.
	d0 := conf.Devices[0].Domains[0].Name
	d1 := conf.Devices[1].Domains[0].Name
	if d0 != d1 {
		return configError("Each domain 0 expected to map to same name. Device 0: %s Device 1: %s", d0, d1)
	}

	return nil
}

func loadOptions(conf ConfigFile) (opts configOptions, err error) {
	opts.fabricCtrlMode = fabricControllerPciMode
	opts.virtMgmtMode = virtualizationManagmentPciMode

	for _, optMap := range conf.Options {
		for key := range optMap {
			val := optMap[key]
			fmt.Printf("Option %s Val %s\n", key, val)
			switch key {
			case "fabric-ctrl":
				switch val {
				case "uart":
					opts.fabricCtrlMode = fabricControllerUartMode
				case "pci":
					opts.fabricCtrlMode = fabricControllerPciMode
				default:
					err = fmt.Errorf("Unsupported fabric-ctrl option %s", val)

				}
			case "virt-mgmt-ctrl":
				switch val {
				case "uart":
					opts.virtMgmtMode = virtualizationManagmentUartMode
				case "pci":
					opts.virtMgmtMode = virtualizationManagmentPciMode
				default:
					err = fmt.Errorf("Unsupport virt-mgmt-ctrl option %s", val)
				}
				// add future options here
			}

			if err != nil {
				break
			}
		}
	}

	fmt.Printf("Options Loaded:\n")
	fmt.Printf("  %-32s : %10s\n", "Fabric Controller", fabricControllerModeName[opts.fabricCtrlMode])
	fmt.Printf("  %-32s : %10s\n", "Virt-Mgmt Controller", virtualizationManagmentModeName[opts.virtMgmtMode])

	return
}

// Locate Device provides the device path for the given PAX ID. Switchtec devices
// are enumerated by the PCIe bus and the /dev/switchtec%d does not map to the
// actual PAX ID (this is espcially true of the eval-kits). So we query each device
// for the PAX ID individually.
func locateDevice(id int32, mode fabricControllerMode) (string, error) {
	debugPrintf := func(format string, a ...interface{}) {
		fmt.Printf("Locate Device: "+format, a...)
	}

	basename := map[fabricControllerMode]string{
		fabricControllerPciMode:  "switchtec",
		fabricControllerUartMode: "ttyUSB",
	}

	for i := 0; i < 10; i++ {
		devPath := path.Join("/dev", fmt.Sprintf("%s%d", basename[mode], i))
		debugPrintf("Stat Device %s\n", devPath)
		if _, err := os.Stat(devPath); os.IsNotExist(err) {
			continue
		}

		debugPrintf("Open %s\n", devPath)
		dev, err := switchtec.Open(devPath)
		if err != nil {
			return "", err
		}
		defer dev.Close()

		debugPrintf("Identify\n")
		devID, err := dev.Identify()
		if err != nil {
			return "", err
		}

		debugPrintf("Device %d\n", devID)
		if devID == id {
			return devPath, nil
		}
	}

	return "", fmt.Errorf("Device Not Found")
}

// locateUartDevice provides the UART device path for the given PAX ID. UART devices
// are enumerated by the OS and so we use a discovery process
func locateUartDevice(id int32) (string, error) {
	debugPrintf := func(format string, a ...interface{}) {
		fmt.Printf("Locate UART Device: ")
		fmt.Printf(format, a...)
	}

	for i := 0; i < 10; i++ {
		devPath := path.Join("/dev", fmt.Sprintf("ttyUSB%d", i))
		debugPrintf("Stat Device %s\n", devPath)
		if _, err := os.Stat(devPath); os.IsNotExist(err) {
			continue
		}

		debugPrintf("Open %s\n", devPath)
		dev, err := switchtec.Open(devPath)
		if err != nil {
			return "", err
		}
		defer dev.Close()

		debugPrintf("Identify device...\n")
		devID, err := dev.Identify()
		if err != nil {
			return "", err
		}

		debugPrintf("Device id %d\n")
		if devID == id {
			return devPath, nil
		}
	}

	return "", fmt.Errorf("UART device not found")
}

type fabricControllerMode int

const (
	fabricControllerUartMode fabricControllerMode = iota
	fabricControllerPciMode
)

var (
	fabricControllerModeName = map[fabricControllerMode]string{
		fabricControllerUartMode: "uart",
		fabricControllerPciMode:  "pci",
	}
)

//
// The Virtualization Management Controller provides an abstraction into
// Virtualization Management command.
//
// Microchip currently requires a side-band approach using Eth, UART, or
// I2C to tunnel into the end-point and submit the commands. We implement
// the UART functionality here as the primary means to issue virt-mgmt
// commands.
//
// In the future this is expected to be PCI driven. Support for this is
// March 2021. Until then... stubs!
//
// See configuration option 'virt-mgmt-ctrl'
type virtualizationManagmentMode int

const (
	virtualizationManagmentUartMode virtualizationManagmentMode = iota
	virtualizationManagmentPciMode
)

var (
	virtualizationManagmentModeName = map[virtualizationManagmentMode]string{
		virtualizationManagmentUartMode: "uart",
		virtualizationManagmentPciMode:  "pci",
	}
)

type virtualizationManagmentController interface {
	list(dev *switchtec.Device, pdfid uint16) (*nvme.SecondaryControllerList, error)
	manage(dev *switchtec.Device, pdfid uint16, controllerID uint16, action nvme.VirtualManagementAction, resourceType nvme.VirtualManagementResourceType, numResources uint32) error
}

type virtualizationManagmentPciController struct {
	nvmeDev *nvme.Device
}

func (ctrl *virtualizationManagmentPciController) list(dev *switchtec.Device, pdfid uint16) (*nvme.SecondaryControllerList, error) {
	return nil, fmt.Errorf("PCI is an unsupported virtualization management controller")
}

func (ctrl *virtualizationManagmentPciController) manage(dev *switchtec.Device, pdfid uint16, controllerID uint16, action nvme.VirtualManagementAction, resourceType nvme.VirtualManagementResourceType, numResources uint32) error {
	return fmt.Errorf("PCI is an unsupported virtualization management controller")
}

type virtualizationManagmentUartController struct {
	dev   *nvme.Device
	devID int32
	pdfid uint16
}

// method to open an end-point tunnel to the device, if necessary
func (ctrl *virtualizationManagmentUartController) openTunnelIfNecessary(dev *switchtec.Device, pdfid uint16) error {
	if ctrl.dev == nil || ctrl.pdfid != pdfid {

		var uartPath string = ""
		var err error = nil

		if ctrl.dev != nil {

			// If the device IDs match, we already know the UART path
			// and don't need to re-locate the device.
			if ctrl.devID == dev.ID() {
				uartPath = ctrl.dev.Path
			}

			ctrl.dev.Close() // TODO: We should be able to just disconnect & re-open the EP tunnel; this is a big handed clobber
		}

		if uartPath == "" {
			uartPath, err = locateUartDevice(dev.ID())
			if err != nil {
				return err
			}
		}

		devPath := fmt.Sprintf("%#04x@%s", pdfid, uartPath)
		ctrl.dev, err = nvme.Open(devPath)
		if err != nil {
			return err
		}

		ctrl.devID = dev.ID()
		ctrl.pdfid = pdfid
	}

	return nil
}

func (ctrl *virtualizationManagmentUartController) list(dev *switchtec.Device, pdfid uint16) (*nvme.SecondaryControllerList, error) {
	if err := ctrl.openTunnelIfNecessary(dev, pdfid); err != nil {
		return nil, err
	}

	return ctrl.dev.ListSecondary(0, 0)
}

func (ctrl *virtualizationManagmentUartController) manage(dev *switchtec.Device, pdfid uint16, controllerID uint16, action nvme.VirtualManagementAction, resourceType nvme.VirtualManagementResourceType, numResources uint32) error {
	if err := ctrl.openTunnelIfNecessary(dev, pdfid); err != nil {
		return err
	}

	return ctrl.dev.VirtualMgmt(controllerID, action, resourceType, numResources)
}

func openVirtualizationManagementController(dev *switchtec.Device, mode virtualizationManagmentMode) (virtualizationManagmentController, error) {
	switch mode {
	case virtualizationManagmentPciMode:
		return &virtualizationManagmentPciController{}, fmt.Errorf("PCI is an unsupported virtualization management controller")
	case virtualizationManagmentUartMode:
		return &virtualizationManagmentUartController{dev: nil}, nil
	}

	return nil, fmt.Errorf("Unsupported virtualization management mode")
}

func configureVirtualizationManagementController(dev *switchtec.Device, ctrl virtualizationManagmentController, pdfid uint16, controllerID uint16, dryRun bool) error {

	fmt.Printf("Retrieve secondary controller info...")
	info, err := getSecondaryControllerInfo(dev, ctrl, pdfid, controllerID)
	if err != nil {
		return err
	}

	fmt.Printf("Secondary Controller Info:\n")
	fmt.Printf("  %-12s: %-32s : %04x\n", "SCID", "Secondary Controller Identifier", info.SecondaryControllerID)
	fmt.Printf("  %-12s: %-32s : %04x\n", "PCID", "Primary Controller Identifier", info.PrimaryControllerID)
	fmt.Printf("  %-12s: %-32s : %04x\n", "SCS", "Secondary Controller State", info.SecondaryControllerState)
	fmt.Printf("  %-12s: %-32s : %04x\n", "VFN", "Virtual Function Number", info.VirtualFunctionNumber)
	fmt.Printf("  %-12s: %-32s : %04x\n", "NVQ", "Num VQ Flex Resources Assigned", info.VQFlexibleResourcesAssigned)
	fmt.Printf("  %-12s: %-32s : %04x\n", "NVI", "Num VI Flex Resources Assigned", info.VIFlexibleResourcesAssigned)

	type assignment struct {
		assigned uint16
		rtype    nvme.VirtualManagementResourceType
	}

	assignments := []assignment{
		{assigned: info.VQFlexibleResourcesAssigned, rtype: nvme.VQResourceType},
		{assigned: info.VIFlexibleResourcesAssigned, rtype: nvme.VIResourceType},
	}

	for _, a := range assignments {
		if a.assigned < 2 {
			fmt.Printf("Assigning %s Flex Resources...", nvme.VirtualManagementResourceTypeName[a.rtype])
			if !dryRun {
				err := ctrl.manage(dev, pdfid, controllerID, nvme.SecondaryAssignAction, a.rtype, 2-uint32(a.assigned))
				if err != nil {
					return err
				}
			}
			fmt.Printf(" Assigned.\n")
		}
	}

	if info.SecondaryControllerState == 0 {
		fmt.Printf("Online Secondary Controller...")

		if !dryRun {
			err := ctrl.manage(dev, pdfid, controllerID, nvme.SecondaryOnlineAction, nvme.VQResourceType, 0)
			if err != nil {
				return err
			}
		}
		fmt.Printf(" Online.\n")
	}

	return nil
}

func getSecondaryControllerInfo(dev *switchtec.Device, ctrl virtualizationManagmentController, pdfid uint16, controllerID uint16) (*nvme.SecondaryControllerEntry, error) {
	list, err := ctrl.list(dev, pdfid)
	if err != nil {
		return nil, err
	}

	for _, info := range list.Entries {
		if info.SecondaryControllerID == controllerID {
			return &info, nil
		}
	}

	return nil, fmt.Errorf("No Secondary Controller Info")
}
