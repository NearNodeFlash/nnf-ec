package cmd

import (
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/switchtec"
)

// IdentifyCmd defines the Identify CLI command and parameters
type IdentifyCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will execute the Identify CLI Command
func (cmd *IdentifyCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		id, err := dev.Identify()

		fmt.Printf("PAX ID: %d\n", id)

		return err
	})
}
