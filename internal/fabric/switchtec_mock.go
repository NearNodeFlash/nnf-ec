package fabric

import (
	"fmt"
	"math/rand"
	"strconv"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
)

type MockSwitchtecController struct {
	devices     []MockSwitchtecDevice
	globalPdfid int
}

type MockSwitchtecDevice struct {
	ctrl *MockSwitchtecController

	id   int
	path string
	open bool

	config *SwitchConfig
	ports  []MockSwitchtecPort
}

type MockSwitchtecPort struct {
	id        int
	functions []switchtec.DumpEpPortAttachedDeviceFunction

	pdfid uint16

	bindings []*switchtec.DumpEpPortAttachedDeviceFunction

	config *PortConfig
}

func NewMockSwitchtecController() SwitchtecControllerInterface {

	if err := loadConfig(); err != nil {
		return nil
	}

	ctrl := &MockSwitchtecController{
		globalPdfid: rand.Intn(0x1000),
	}

	ctrl.devices = make([]MockSwitchtecDevice, len(Config.Switches))
	for switchIdx, switchConfig := range Config.Switches {

		devId, err := strconv.Atoi(switchConfig.Id)
		if err != nil {
			return nil
		}

		dev := MockSwitchtecDevice{
			ctrl:   ctrl,
			id:     devId,
			path:   "",
			config: &Config.Switches[switchIdx],
		}

		ctrl.devices[switchIdx] = dev

		dev.ports = make([]MockSwitchtecPort, len(switchConfig.Ports))
		for portIdx, portConfig := range switchConfig.Ports {

			port := MockSwitchtecPort{
				id:     portIdx,
				pdfid:  uint16(ctrl.allocateNewPDFID()),
				config: &switchConfig.Ports[portIdx],
			}

			switch portConfig.getPortType() {
			case sf.MANAGEMENT_PORT_PV130PT:
				port.bindings = make([]*switchtec.DumpEpPortAttachedDeviceFunction, switchConfig.DownstreamPortCount)
			case sf.UPSTREAM_PORT_PV130PT:
				port.bindings = make([]*switchtec.DumpEpPortAttachedDeviceFunction, Config.DownstreamPortCount)
			case sf.DOWNSTREAM_PORT_PV130PT:
				{
					port.functions = make([]switchtec.DumpEpPortAttachedDeviceFunction, 1 /* PF */ +1 /* MGMT */ +Config.UpstreamPortCount)

					isPF := func(idx int) uint8 {
						if idx == 0 {
							return 1
						}
						return 0
					}

					for idx := range port.functions {

						f := switchtec.DumpEpPortAttachedDeviceFunction{
							FunctionID:     uint16(idx),
							PDFID:          port.pdfid + uint16(idx),
							SRIOVCapPF:     isPF(idx),
							VFNum:          uint8(idx - 1),
							Bound:          0,
							BoundPAXID:     0,
							BoundHVDPhyPID: 0,
							BoundHVDLogPID: 0,
						}

						port.functions[idx] = f

					}
				}
			}

			dev.ports[portIdx] = port
		}
	}

	return ctrl
}

func (c MockSwitchtecController) Exists(path string) bool { return true } // TODO: Some sort of testing where one or more switchtec device is missing

func (c MockSwitchtecController) Open(path string) (SwitchtecDeviceInterface, error) {
	for deviceIdx := range c.devices {
		device := &c.devices[deviceIdx]
		if device.path == "" || device.path == path {
			device.path = path
			return device, nil
		}
	}
	return nil, fmt.Errorf("Device %s not found", path)
}

func (c *MockSwitchtecController) allocateNewPDFID() int {
	pdfid := c.globalPdfid
	c.globalPdfid += 0x100
	return pdfid
}

func (d MockSwitchtecDevice) Close() {
	d.path = ""
	d.id = -1
}

func (d MockSwitchtecDevice) Identify() (int32, error) {
	return int32(d.id), nil
}

func (d MockSwitchtecDevice) EnumerateEndpoint(id uint8, handlerFunc func(epPort *switchtec.DumpEpPortDevice) error) error {

	for _, port := range d.ports {
		if uint8(port.id) == id {
			epPort := switchtec.DumpEpPortDevice{
				Ep: switchtec.DumpEpPortEp{
					Functions: port.functions,
				},
			}

			return handlerFunc(&epPort)
		}
	}

	return nil
}

func (d MockSwitchtecDevice) Bind(hostPhysPortId, hostLogPortId uint8, pdfid uint16) error {

	bindPort := func(hostPort *MockSwitchtecPort, hostLogPortId uint8, pdfid uint16) error {
		for _, port := range d.ports {

			if port.pdfid == pdfid&0xFF00 {
				for functionIdx := range port.functions {
					function := &port.functions[functionIdx]

					if function.PDFID == pdfid {
						if function.Bound != 0 {
							return fmt.Errorf("Device %#04x already bound", pdfid)
						}

						function.Bound = 1
						function.BoundHVDPhyPID = hostPhysPortId
						function.BoundHVDLogPID = hostLogPortId
						function.BoundPAXID = uint8(d.id)

						hostPort.bindings[hostLogPortId] = function
					}
				}
			}
		}

		return fmt.Errorf("PDFID %#04x Not Found", pdfid)
	}

	for portIdx, port := range d.ports {
		if port.config.Port == int(hostPhysPortId) {
			if port.bindings[hostLogPortId] != nil {
				return fmt.Errorf("Host Port %d Logical Port ID %d already bound", hostPhysPortId, hostLogPortId)
			}

			return bindPort(&d.ports[portIdx], hostLogPortId, pdfid)
		}
	}

	return fmt.Errorf("Host Physical Port ID %d not found", hostPhysPortId)
}
