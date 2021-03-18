package nvmenamespace

import (
	"encoding/binary"
	"fmt"
	"math"

	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
)

type NvmeMockController struct{}

func NewMockNvmeController() NvmeControllerInterface {
	return &NvmeMockController{}
}

// NewNvmeStorageController - Interface into communicating with an NVMe Storage device
func (*NvmeMockController) NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceInterface, error) {
	return NewMockDevice(), nil
}

const (
	invalidNamespaceId = nvme.NamespaceIdentifier(0)

	mockSecondaryControllerCount  = 17
	mockMaximumNamespaceCount     = 32
	mockControllerCapacityInBytes = 2 << 40 // 2TiB
)

type NvmeMockDevice struct {
	virtualizationManagement bool
	controllers              [1 + mockSecondaryControllerCount]NvmeMockDeviceController
}

func NewMockDevice() *NvmeMockDevice {
	mock := NvmeMockDevice{
		virtualizationManagement: true,
	}

	for idx := range mock.controllers {
		mock.controllers[idx] = NvmeMockDeviceController{
			id:          uint16(idx),
			capacity:    mockControllerCapacityInBytes,
			online:      false,
			vqresources: 0,
			viresources: 0,
		}
	}

	return &mock
}

// These structures represent the objects existing on an NVMe drive

type NvmeMockDeviceController struct {
	id                uint16
	online            bool
	capacity          uint64
	allocatedCapacity uint64
	vqresources       uint16
	viresources       uint16
	namespaces        [mockMaximumNamespaceCount]NvmeMockNamespace
}

type NvmeMockNamespace struct {
	id                  nvme.NamespaceIdentifier
	size                uint64
	capacity            uint64
	guid                [16]byte
	attachedControllers [mockSecondaryControllerCount]*NvmeMockDeviceController
}

// NewNvmeDeviceController -
func (d *NvmeMockDevice) NewNvmeDeviceController(controllerId uint16) NvmeDeviceControllerInterface {
	for idx := range d.controllers {
		if d.controllers[idx].id == controllerId {
			return &d.controllers[idx]
		}
	}

	return nil
}

// IdentifyController -
func (d *NvmeMockDevice) IdentifyController() (*nvme.IdCtrl, error) {
	ctrl := new(nvme.IdCtrl)

	binary.LittleEndian.PutUint64(ctrl.TotalNVMCapacity[:], mockControllerCapacityInBytes)
	binary.LittleEndian.PutUint64(ctrl.UnallocatedNVMCapacity[:], mockControllerCapacityInBytes)

	ctrl.OptionalAdminCommandSupport = nvme.VirtualiztionManagementSupport

	return ctrl, nil
}

// IdentifyNamespace -
func (d *NvmeMockDevice) IdentifyNamespace() (*nvme.IdNs, error) {
	ns := new(nvme.IdNs)

	ns.NumberOfLBAFormats = 1
	ns.LBAFormats[0].LBADataSize = uint8(math.Log2(4096))
	ns.LBAFormats[0].MetadataSize = 0
	ns.LBAFormats[0].RelativePerformance = 0

	return ns, nil
}

// EnumerateSecondaryControllers -
func (d *NvmeMockDevice) EnumerateSecondaryControllers(initFunc SecondaryControllersInitFunc, handlerFunc SecondaryControllerHandlerFunc) error {
	initFunc(uint8(mockSecondaryControllerCount))

	for _, ctrl := range d.controllers {
		handlerFunc(ctrl.id, ctrl.online, ctrl.id, ctrl.vqresources, ctrl.viresources)
	}

	return nil
}

// AssignControllerResources -
func (d *NvmeMockDevice) AssignControllerResources(controllerId uint16, resourceType SecondaryControllerResourceType, numResources uint32) error {
	ctrl := &d.controllers[int(controllerId)]
	switch resourceType {
	case VQResourceType:
		ctrl.vqresources = uint16(numResources)
	case VIResourceType:
		ctrl.viresources = uint16(numResources)
	}
	return nil
}

// OnlineController -
func (d *NvmeMockDevice) OnlineController(controllerId uint16) error {
	ctrl := &d.controllers[int(controllerId)]
	ctrl.online = true

	return nil
}

// ListNamespaces -
func (d *NvmeMockDevice) ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error) {
	nss := d.controllers[controllerId].namespaces
	var count = 0
	for _, ns := range nss {
		if ns.id != 0 {
			count++
		}
	}

	list := make([]nvme.NamespaceIdentifier, count)
	for idx, ns := range nss {
		if ns.id != invalidNamespaceId {
			list[idx] = ns.id
		}
	}

	return list, nil
}

// GetNamespace -
func (d *NvmeMockDevice) GetNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error) {
	if namespaceId == invalidNamespaceId {
		return nil, fmt.Errorf("Namespace not found")
	}

	for _, ns := range d.controllers[0].namespaces {
		if ns.id == namespaceId {
			nsid := &nvme.IdNs{
				Size:                     uint64(ns.size),
				Capacity:                 uint64(ns.capacity),
				GloballyUniqueIdentifier: ns.guid,
				FormattedLBASize: nvme.FormattedLBASize{
					Format: 0,
				},
				MultiPathIOSharingCapabilities: nvme.NamespaceCapabilities{
					Sharing: 1,
				},
			}

			nsid.NumberOfLBAFormats = 1
			nsid.LBAFormats[0].LBADataSize = uint8(math.Log2(4096))
			nsid.LBAFormats[0].MetadataSize = 0

			return nsid, nil
		}
	}

	return nil, fmt.Errorf("Namespace not found")
}

// CreateNamespace -
func (d *NvmeMockDevice) CreateNamespace(capacityBytes uint64, metadata []byte) (nvme.NamespaceIdentifier, error) {
	ctrl := &d.controllers[0]

	if capacityBytes > (ctrl.capacity - ctrl.allocatedCapacity) {
		return 0, fmt.Errorf("Insufficient Capacity")
	}

	var id = invalidNamespaceId
	for idx, ns := range ctrl.namespaces {
		if ns.id == invalidNamespaceId {
			id = nvme.NamespaceIdentifier(idx + 1)
			break
		}
	}

	if id == invalidNamespaceId {
		return id, fmt.Errorf("Could not find free namespace")
	}

	mockNs := NvmeMockNamespace{
		id:       id,
		size:     capacityBytes, // TODO: This should be converted to FLBA Size
		capacity: capacityBytes, // TODO: This should be converted to FLBA Size
		guid: [16]byte{
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			byte(id >> 12),
			byte(id >> 8),
			byte(id >> 4),
			byte(id >> 0),
		},
	}

	for idx := range mockNs.attachedControllers {
		mockNs.attachedControllers[idx] = nil
	}

	ctrl.allocatedCapacity += capacityBytes

	for idx, ns := range ctrl.namespaces {
		if ns.id == invalidNamespaceId {
			ctrl.namespaces[idx] = mockNs
			break
		}
	}

	return mockNs.id, nil
}

// DeleteNamespace -
func (d *NvmeMockDevice) DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error {
	if namespaceId == invalidNamespaceId {
		return fmt.Errorf("Namespace %d Not Found", namespaceId)
	}

	ctrl := &d.controllers[0]

	for idx, ns := range ctrl.namespaces {
		if ns.id == namespaceId {

			controllerIds := make([]uint16, len(ns.attachedControllers))
			for controllerIdx, controller := range ns.attachedControllers {
				controllerIds[controllerIdx] = controller.id
			}

			if err := d.DetachNamespace(namespaceId, controllerIds); err != nil {
				return err
			}

			ctrl.allocatedCapacity -= ns.capacity

			ctrl.namespaces[idx].id = 0

			return nil
		}
	}

	return fmt.Errorf("Namespace %d Not Found", namespaceId)
}

// AttachNamespace -
func (d *NvmeMockDevice) AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	if namespaceId == invalidNamespaceId {
		return fmt.Errorf("Namespace %d Not Found", namespaceId)
	}

	ctrl := &d.controllers[0]
	for nsIdx, ns := range ctrl.namespaces {
		if ns.id == namespaceId {
			for _, controller := range controllers {
				found := false
				for ctrlIdx, ctrl := range d.controllers {
					if ctrl.id == controller {
						ctrl.namespaces[nsIdx].attachedControllers[controller] =
							&d.controllers[ctrlIdx]

						found = true
					}
				}

				if !found {
					return fmt.Errorf("Controller %d Not Found", controller)
				}
			}

			break
		}
	}

	return nil
}

// DetachNamespace -
func (d *NvmeMockDevice) DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	if namespaceId == invalidNamespaceId {
		return fmt.Errorf("Namespace %d Not Found", namespaceId)
	}

	ctrl := &d.controllers[0]
	for nsIdx, ns := range ctrl.namespaces {
		if ns.id == namespaceId {
			for _, controller := range controllers {

				found := false
				for ctrlIdx, ctrl := range ns.attachedControllers {
					if ctrl.id == controller {
						ctrl.namespaces[nsIdx].attachedControllers[ctrlIdx] = nil

						found = true
					}
				}

				if !found {
					return fmt.Errorf("Controller %d Not Found", controller)
				}
			}

			break
		}
	}

	return nil
}

// ListNamespaces -
func (c *NvmeMockDeviceController) ListNamespaces() ([]nvme.NamespaceIdentifier, error) {
	nsids := []nvme.NamespaceIdentifier{}
	for idx, ns := range c.namespaces {
		nsids[idx] = ns.id
	}
	return nsids, nil
}
