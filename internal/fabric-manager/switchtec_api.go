package fabric

import (
	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
)

type SwitchtecControllerInterface interface {
	Open(path string) (SwitchtecDeviceInterface, error)
}

type SwitchtecDeviceInterface interface {
	Close()

	Identify() (int32, error)

	GetFirmwareVersion() (string, error)
	GetModel() (string, error)
	GetManufacturer() (string, error)
	GetSerialNumber() (string, error)

	EnumerateEndpoint(uint8, func(epPort *switchtec.DumpEpPortDevice) error) error

	Bind(uint8, uint8, uint16) error
}
