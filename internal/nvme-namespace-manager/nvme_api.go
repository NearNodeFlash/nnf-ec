package nvmenamespace

import (
	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
)

// NvmeControllerInterface -
type NvmeControllerInterface interface {
	NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceInterface, error)
}

// SecondaryControllersInitFunc -
type SecondaryControllersInitFunc func(count uint8)

// SecondaryControllerHandlerFunc -
type SecondaryControllerHandlerFunc func(controllerId uint16, controllerOnline bool, virtualFunctionNumber uint16, numVQResourcesAssinged, numVIResourcesAssigned uint16) error

// SecondaryControllerResourceType -
type SecondaryControllerResourceType int

const (
	VQResourceType SecondaryControllerResourceType = iota
	VIResourceType
)

type VolumeId uint32

// NvmeDeviceInterface -
type NvmeDeviceInterface interface {
	NewNvmeDeviceController(controllerId uint16) NvmeDeviceControllerInterface

	IdentifyController() (*nvme.IdCtrl, error)
	IdentifyNamespace() (*nvme.IdNs, error)

	EnumerateSecondaryControllers(
		SecondaryControllersInitFunc,
		SecondaryControllerHandlerFunc) error

	AssignControllerResources(
		controllerId uint16,
		resourceType SecondaryControllerResourceType,
		numResources uint32) error

	OnlineController(controllerId uint16) error

	ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error)

	GetNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error)

	CreateNamespace(sizeInBytes uint64) (nvme.NamespaceIdentifier, error)
	DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error

	AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllerId uint16) error
	DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllerId uint16) error
}

// NvmeDeviceControllerInterface -
type NvmeDeviceControllerInterface interface {
	ListNamespaces() ([]nvme.NamespaceIdentifier, error)
}
