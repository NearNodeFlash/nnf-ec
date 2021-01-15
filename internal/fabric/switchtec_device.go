package fabric

import (
	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
)

type SwitchtecController struct{}

func NewSwitchtecController() SwitchtecControllerInterface {
	return &SwitchtecController{}
}

func (SwitchtecController) Open(path string) (SwitchtecDeviceInterface, error) {
	dev, err := switchtec.Open(path)
	return &SwitchtecDevice{dev: dev}, err
}

type SwitchtecDevice struct {
	dev *switchtec.Device
}

func (d SwitchtecDevice) Close() {
	d.dev.Close()
}

func (d SwitchtecDevice) Identify() (int32, error) {
	return d.dev.Identify()
}

func (d SwitchtecDevice) EnumerateEndpoint(id uint8, f func(epPort *switchtec.DumpEpPortDevice) error) error {
	return d.dev.GfmsEpPortDeviceEnumerate(id, f)
}

func (d SwitchtecDevice) Bind(hostPhysPortId, hostLogPortId uint8, pdfid uint16) error {
	return d.dev.Bind(uint8(d.dev.ID()), hostPhysPortId, hostLogPortId, pdfid)
}
