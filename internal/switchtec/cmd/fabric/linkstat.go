package cmd

import (
	"fmt"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

// LinkStatCmd defines the Link Stat CLI command and parameters
type LinkStatCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will execute the Link Stat CLI Command
func (cmd *LinkStatCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		stats, err := dev.LinkStat()
		if err != nil {
			return err
		}

		fmt.Printf("Link Stats\n")
		for _, s := range stats {
			fmt.Printf("Physical Port ID: %d\n", s.PhysPortId)

			fmt.Printf("  %-32s : %v\n", "Link Up", s.LinkUp)
			fmt.Printf("  %-32s : Gen%d\n", "Link Gen", s.LinkGen)
			fmt.Printf("  %-32s : %04x\n", "Link State", uint16(s.LinkState))

			fmt.Printf("  %-32s : x%d\n", "Configured Link Width", s.CfgLinkWidth)
			fmt.Printf("  %-32s : x%d\n", "Negotiated Link Width", s.NegLinkWidth)

			fmt.Printf("  %-32s : %4.2f GBps\n", "Current Link Rate", s.CurLinkRateGBps)
		}

		return nil
	})
}
