package nvmenamespace

import (
	. "stash.us.cray.com/rabsw/nnf-ec/internal/common"

	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/nvme"
)

type NvmeController struct {
}

func NewNvmeController() NvmeControllerInterface {
	return &NvmeController{}
}

func (NvmeController) NewNvmeStorageController(fabricId, switchId, portId string) (NvmeStorageControllerInterface, error) {
	d, err := FabricController.GetSwitchtecDevice(fabricId, switchId)
	if err != nil {
		return nil, err
	}

	pdfid, err := FabricController.GetPortPDFID(fabricId, switchId, portId)
	if err != nil {
		return nil, err
	}

	dev, err := nvme.Connect(d, pdfid)

	return &NvmeStorageController{dev: dev}, nil
}

type NvmeStorageController struct {
	dev *nvme.Device
}

func (c NvmeStorageController) Identify() {

}
