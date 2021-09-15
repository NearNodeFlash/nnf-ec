package cmd

import (
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/switchtec"
)

// GetSerialCmd -
type GetSerialCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will run the GetSerialCmd command
func (cmd *GetSerialCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		serial, err := dev.GetSerialNumber()

		fmt.Printf("Device Serial Number: %#08x\n", serial)
		return err
	})
}

// GetFirmwareCmd -
type GetFirmwareCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will run the GetFirmwareCmd command
func (cmd *GetFirmwareCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		ver, err := dev.GetFirmwareVersion()

		fmt.Printf("Device Firmware: %s\n", ver)

		return err
	})
}

type MfgCmd struct {
	Serial   GetSerialCmd   `kong:"cmd,help='Retrieve the device serial number.'"`
	Firmware GetFirmwareCmd `kong:"cmd,help='Retrieve the active firmware.'"`
}
