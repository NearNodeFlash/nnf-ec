package nvmenamespace

// NvmeControllerInterface -
type NvmeControllerInterface interface {
	NewNvmeStorageController(fabricId, switchId, portId string) (NvmeStorageControllerInterface, error)
}


// NvmeStorageController -
type NvmeStorageControllerInterface interface {
	Identify()

	//ListSecondary() (secondaryControllers, error)

	/*
		Manage()

		Create()
		Delete()
		Attach()
		Detach()
	*/
}
