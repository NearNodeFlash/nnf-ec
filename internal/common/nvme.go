package common

type NvmeApi interface {
	GetVolumes(controllerId string) ([]string, error)
	AttachVolume(odataid string, controllerId string) error
}

var NvmeInterface NvmeApi

func RegisterNvmeInterface(api NvmeApi) {
	NvmeInterface = api
}
