package nvmenamespace

import (
	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
)

func NewCliNvmeController() NvmeController {
	return &cliNvmeController{}
}

type cliNvmeController struct {}

func (cliNvmeController) NewNvmeDeviceController() NvmeDeviceController {
	return &cliNvmeDeviceController{}
}

type cliNvmeDeviceController struct{}

func (cliNvmeDeviceController) NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceApi, error) {
	return &cliDevice{}, nil
}

type cliDevice struct {
	path  string // Path to owning switchtec device
	pdfid uint16 // PDFID of the device
}

func (*cliDevice) IdentifyController(controllerId uint16) (*nvme.IdCtrl, error) {
	return nil, nil
}

func (*cliDevice) IdentifyNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error) {
	return nil, nil
}

func (*cliDevice) EnumerateSecondaryControllers(initFunc SecondaryControllersInitFunc, handlerFunc SecondaryControllerHandlerFunc) error {
	return nil
}

func (*cliDevice) AssignControllerResources(controllerId uint16, resourceType SecondaryControllerResourceType, numResources uint32) error {
	return nil
}

func (*cliDevice) OnlineController(controllerId uint16) error {
	return nil
}

func (*cliDevice) ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error) {
	return nil, nil
}

func (*cliDevice) CreateNamespace(capacityBytes uint64, metadata []byte) (nvme.NamespaceIdentifier, error) {
	return ^nvme.NamespaceIdentifier(0), nil
}

func (*cliDevice) DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error {
	return nil
}

func (*cliDevice) AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	return nil
}

func (*cliDevice) DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error {
	return nil
}
