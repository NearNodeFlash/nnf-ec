package api

type NvmeApi interface {
	GetVolumes(controllerId string) ([]string, error)
}

var NvmeInterface NvmeApi

func RegisterNvmeInterface(api NvmeApi) {
	NvmeInterface = api
}
