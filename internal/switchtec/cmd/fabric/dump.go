package cmd

import (
	"fmt"

	"github.com/HewlettPackard/structex"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

type DumpEpPortCmd struct {
	Device string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Pid    uint8  `arg help:"The end-point port ID."`
}

func (cmd *DumpEpPortCmd) Run() error {

	fmt.Printf("%s Dump EP Port %d\n", cmd.Device, cmd.Pid)

	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	fmt.Printf("PAX %d Device Opened. Port %d Start...\n", dev.ID(), cmd.Pid)

	len, err := dev.GfmsEpPortStart(cmd.Pid)
	if err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Started. Port Get %d dwords...\n", len)

	buf, err := dev.GfmsEpPortGet(cmd.Pid, len)
	if err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Buf %d bytes \n%v\n", buf.Len(), buf)
	fmt.Printf("Port Finish...\n")

	if err := dev.GfmsEpPortFinish(); err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Finished\n")

	ep := new(switchtec.DumpEpPortDevice)

	if true {
		for i := 0; i < buf.Len(); i++ {
			if i%16 == 0 {
				fmt.Printf("\n%08x: ", i)
			}
			if i%8 == 0 {
				fmt.Printf(" ")
			}
			fmt.Printf("%02x", buf.Bytes()[i])
		}

		fmt.Printf("\n")
	}

	if err := structex.DecodeByteBuffer(buf, ep); err != nil {
		return err
	}

	fmt.Printf("%+v\n", ep)

	return nil
}

type DumpCmd struct {
	EpPort DumpEpPortCmd `cmd help:"Dump information for a specific end-point port."`
}
