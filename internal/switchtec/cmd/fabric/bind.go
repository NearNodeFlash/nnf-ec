package cmd

import (
	"strconv"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

// BindCmd defines the Bind CLI command and parameters
type BindCmd struct {
	Device    string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	SwIndex   uint8  `arg help:"Software index of the fabric device."`
	PhyPortID uint8  `arg help:"Physical port ID within the domain."`
	LogPortID uint8  `arg help:"Logical port ID within the domain."`
	PDFID     string `arg help:"PDFID of the end-point."`
}

// Run will run the Bind command
func (cmd *BindCmd) Run() error {
	pdfid, err := strconv.ParseUint(cmd.PDFID, 0, 16)
	if err != nil {
		return err
	}

	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return dev.Bind(cmd.SwIndex, cmd.PhyPortID, cmd.LogPortID, uint16(pdfid))
}

// UnbindCmd defines the Bind CLI command and parameters
type UnbindCmd struct {
	Device    string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	SwIndex   uint8  `arg help:"Software index of the fabric device."`
	PhyPortID uint8  `arg help:"Physical port ID within the domain."`
	LogPortID uint8  `arg help:"Logical port ID within the domain."`
}

// Run will run the Unbind command
func (cmd *UnbindCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}

	return dev.Unbind(cmd.SwIndex, cmd.PhyPortID, cmd.LogPortID)
}
