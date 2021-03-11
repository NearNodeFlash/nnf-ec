package nvmenamespace

import (
	"stash.us.cray.com/rabsw/nnf-ec/internal/api"

	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/switchtec"
)

type NvmeController struct {
}

func NewNvmeController() NvmeControllerInterface {
	return &NvmeController{}
}

func (*NvmeController) NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceInterface, error) {
	switchDev, err := api.FabricController.GetSwitchtecDevice(fabricId, switchId)
	if err != nil {
		return nil, err
	}

	pdfid, err := api.FabricController.GetPortPDFID(fabricId, switchId, portId, 0)
	if err != nil {
		return nil, err
	}

	nvmeDev, err := nvme.Connect(switchDev, pdfid)

	d := NvmeDevice{
		fabricId:  fabricId,
		switchId:  switchId,
		portId:    portId,
		dev:       nvmeDev,
		switchDev: switchDev,
	}

	return &d, nil
}

type NvmeDevice struct {
	fabricId string
	switchId string
	portId   string

	dev       *nvme.Device // Represents the PF of the device
	switchDev *switchtec.Device
}

func (d *NvmeDevice) NewNvmeDeviceController(controllerId uint16) NvmeDeviceControllerInterface {
	ctrl := &NvmeDeviceController{
		controllerId: controllerId,
		parent:       d,
		dev:          nil,
	}

	return ctrl
}

// IdentifyController -
func (d *NvmeDevice) IdentifyController() (*nvme.IdCtrl, error) {
	return d.dev.IdentifyController()
}

// IdentifyNamespace -
func (d *NvmeDevice) IdentifyNamespace() (*nvme.IdNs, error) {
	return d.dev.IdentifyNamespace(uint32(nvme.COMMON_NAMESPACE_IDENTIFIER), false)
}

// EnumerateSecondaryControllers -
func (d *NvmeDevice) EnumerateSecondaryControllers(initFunc SecondaryControllersInitFunc, handlerFunc SecondaryControllerHandlerFunc) error {

	list, err := d.dev.ListSecondary(0, 0)
	if err != nil {
		return err
	}

	initFunc(list.Count)

	count := int(list.Count)
	for i := 0; i < count; i++ {
		entry := list.Entries[i]

		err := handlerFunc(
			entry.SecondaryControllerID,
			entry.SecondaryControllerState&0x1 != 0,
			entry.VirtualFunctionNumber,
			entry.VQFlexibleResourcesAssigned,
			entry.VIFlexibleResourcesAssigned,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// AssignControllerResources -
func (d *NvmeDevice) AssignControllerResources(controllerId uint16, resourceType SecondaryControllerResourceType, numResources uint32) error {

	if err := d.dev.VirtualMgmt(controllerId, nvme.SecondaryAssignAction, nvme.VQResourceType, numResources); err != nil {
		return err
	}

	if err := d.dev.VirtualMgmt(controllerId, nvme.SecondaryAssignAction, nvme.VIResourceType, numResources); err != nil {
		return err
	}

	return nil
}

// OnlineController -
func (d *NvmeDevice) OnlineController(controllerId uint16) error {
	return d.dev.VirtualMgmt(controllerId, nvme.SecondaryOnlineAction, nvme.VQResourceType /*Ignored for OnlineAction*/, 0 /*Ignored for OnlineAction*/)
}

// ListNamespaces -
func (d *NvmeDevice) ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error) {
	list, err := d.dev.IdentifyNamespaceList(0, true)
	if err != nil {
		return nil, err
	}

	// Compress the returned list to only those IDs which are valid (non-zero)
	ret := make([]nvme.NamespaceIdentifier, len(list))
	var count = 0
	for _, id := range list {
		if id != 0 {
			ret[count] = id
			count++
		}
	}

	return ret[:count], nil
}

// GetNamespace -
func (d *NvmeDevice) GetNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error) {
	return d.dev.IdentifyNamespace(uint32(namespaceId), true)
}

// CreateNamespace -
func (d *NvmeDevice) CreateNamespace(capacityBytes uint64) (nvme.NamespaceIdentifier, error) {

	// Want to get the best LBA format for creating a Namespace
	// We first read the unique namespace ID that describes common namespace properties
	dns, err := d.dev.IdentifyNamespace(uint32(nvme.COMMON_NAMESPACE_IDENTIFIER), false)
	if err != nil {
		return 0, err
	}

	// We then iterate over the LBA formats presented by the drive and look for
	// the best performing LBA format that has no metadata.
	var bestPerformance = ^uint8(0) // Performance improves as the RelativePerformance value gets lower
	var bestIndex = 0
	for i := 0; i < int(dns.NumberOfLBAFormats); i++ {
		if dns.LBAFormats[i].MetadataSize != 0 {
			continue
		}
		if dns.LBAFormats[i].RelativePerformance < bestPerformance {
			bestIndex = i
			bestPerformance = dns.LBAFormats[i].RelativePerformance
		}
	}

	// TODO: We should probably do the above only once when identifying the drive
	// and then check at certain points the requested CapacityBytes is a good
	// value.

	roundUpToMultiple := func(n, m uint64) uint64 {
		return ((n + m - 1) / m) * m
	}

	dataSizeBytes := uint64(1 << dns.LBAFormats[bestIndex].LBADataSize)
	size := roundUpToMultiple(capacityBytes/dataSizeBytes, dataSizeBytes)

	id, err := d.dev.CreateNamespace(
		size,             // Size in Data Size Units (usually 4096)
		size,             // Capacity in Data Size Units (usually 4096),
		uint8(bestIndex), // LBA Format Index (see above)
		0,                // Data Protection Capaiblities (none)
		0x1,              // Capabilities (sharing = 1b)
		0,                // ANA Group Identifier (none)
		0,                // NVM Set Identifier (non)
		100,              // Timeout (???)
	)

	return nvme.NamespaceIdentifier(id), err
}

// DeleteNamespace -
func (d *NvmeDevice) DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error {
	return d.dev.DeleteNamespace(uint32(namespaceId))
}

// AttachNamespace -
func (d *NvmeDevice) AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllerId uint16) error {
	ctrls := [1]uint16{controllerId}
	return d.dev.AttachNamespace(uint32(namespaceId), ctrls[:])
}

// DetachNamespace -
func (d *NvmeDevice) DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllerId uint16) error {
	ctrls := [1]uint16{controllerId}
	return d.dev.DetachNamespace(uint32(namespaceId), ctrls[:])
}

// NvmeDeviceController -
type NvmeDeviceController struct {
	controllerId uint16
	parent       *NvmeDevice
	dev          *nvme.Device
}

// ListNamespaces -
func (c *NvmeDeviceController) ListNamespaces() ([]nvme.NamespaceIdentifier, error) {
	if c.dev == nil {
		pdfid, err := api.FabricController.GetPortPDFID(c.parent.fabricId, c.parent.switchId, c.parent.portId, c.controllerId)
		if err != nil {
			return nil, err
		}

		dev, err := nvme.Connect(c.parent.switchDev, pdfid)
		if err != nil {
			return nil, err
		}

		c.dev = dev
	}

	list, err := c.dev.IdentifyNamespaceList(0, true)
	if err != nil {
		return nil, err
	}

	// Compress the returned list to only those IDs which are valid (non-zero)
	ret := make([]nvme.NamespaceIdentifier, len(list))
	var count = 0
	for _, id := range list {
		if id != 0 {
			ret[count] = id
			count++
		}
	}

	return ret[:count], nil
}
