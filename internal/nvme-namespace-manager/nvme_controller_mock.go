package nvmenamespace

type NvmeMockController struct{}

func NewMockNvmeController() NvmeControllerInterface {
	return &NvmeMockController{}
}

func (NvmeMockController) NewNvmeStorageController(fabricId, switchId, portId string) (NvmeStorageControllerInterface, error) {
	return &NvmeStorageController{}, nil
}

type NvmeMockStorageController struct{}

func (NvmeMockStorageController) Identify() {

}
