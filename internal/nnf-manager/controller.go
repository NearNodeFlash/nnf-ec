package nnf

type NnfControllerInterface interface{}

type NnfController struct{}

func NewNnfController() NnfControllerInterface {
	return &NnfController{}
}

// TODO: See if mock makes sense here - really the lower-layers should provide the mocking
// of physical devices; NNF Controller might be fine building off that.
func NewMockNnfController() NnfControllerInterface {
	return &NnfController{}
}
