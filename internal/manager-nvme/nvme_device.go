package nvme

import (
	"fmt"

	fabric "stash.us.cray.com/rabsw/nnf-ec/internal/manager-fabric"

	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
)

type SwitchtecNvmeController struct{}

func NewSwitchtecNvmeController() NvmeController {
	return &SwitchtecNvmeController{}
}

func (SwitchtecNvmeController) NewNvmeDeviceController() NvmeDeviceController {
	return &SwitchtecNvmeDeviceController{}
}

type SwitchtecNvmeDeviceController struct{}

func (SwitchtecNvmeDeviceController) NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceApi, error) {
	return newNvmeDevice(fabricId, switchId, portId)
}

type nvmeDevice struct {
	dev   *nvme.Device
	pdfid uint16
}

func newNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceApi, error) {
	sdev := fabric.GetSwitchDevice(fabricId, switchId)
	if sdev == nil {
		return nil, fmt.Errorf("Device not found")
	}

	pdfid, err := fabric.GetPortPDFID(fabricId, switchId, portId, 0)
	if err != nil {
		return nil, err
	}

	dev, err := nvme.Connect(sdev, pdfid)
	if err != nil {
		return nil, err
	}

	return &nvmeDevice{dev: dev, pdfid: pdfid}, nil
}

// IdentifyController -
func (d *nvmeDevice) IdentifyController(controllerId uint16) (*nvme.IdCtrl, error) {
	return d.dev.IdentifyController()
}

// IdentifyNamespace -
func (d *nvmeDevice) IdentifyNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error) {
	return d.dev.IdentifyNamespace(uint32(namespaceId), false)
}

// ListSecondary -
func (d *nvmeDevice) ListSecondary() (*nvme.SecondaryControllerList, error) {
	return d.dev.ListSecondary(0, 0)
}

// AssignControllerResources -
func (d *nvmeDevice) AssignControllerResources(controllerId uint16, resourceType SecondaryControllerResourceType, numResources uint32) error {
	resourceTypeMap := map[SecondaryControllerResourceType]nvme.VirtualManagementResourceType{
		VQResourceType: nvme.VQResourceType,
		VIResourceType: nvme.VIResourceType,
	}

	return d.dev.VirtualMgmt(controllerId, nvme.SecondaryAssignAction, resourceTypeMap[resourceType], numResources)
}

// OnlineController -
func (d *nvmeDevice) OnlineController(controllerId uint16) error {
	return d.dev.VirtualMgmt(controllerId, nvme.SecondaryOnlineAction, nvme.VQResourceType /*Ignored for OnlineAction*/, 0 /*Ignored for OnlineAction*/)
}

// ListNamespaces -
func (d *nvmeDevice) ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error) {
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
func (d *nvmeDevice) GetNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error) {
	return d.dev.IdentifyNamespace(uint32(namespaceId), true)
}

// CreateNamespace -
func (d *nvmeDevice) CreateNamespace(capacityBytes uint64, metadata []byte) (nvme.NamespaceIdentifier, error) {

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
func (d *nvmeDevice) DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error {
	return d.dev.DeleteNamespace(uint32(namespaceId))
}

// AttachNamespace -
func (d *nvmeDevice) AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	return d.dev.AttachNamespace(uint32(namespaceId), controllers)
}

// DetachNamespace -
func (d *nvmeDevice) DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	return d.dev.DetachNamespace(uint32(namespaceId), controllers)
}
