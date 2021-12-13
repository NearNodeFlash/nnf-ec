package cmd

import (
	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

// EchoCmd defines the Echo CLI command and parameters
type EchoCmd struct {
	Device  string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Payload uint32 `arg optional default:"0" help:"The echo payload. The bit-inverse will be returned by the device."`
}

// Run will run the Echo command
func (cmd *EchoCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return dev.Echo(cmd.Payload)
}
