package nvme

import (
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/nvme"
)

// CreateNsCmd sends a namespace management command to the specified
// device to create a namespace with the given parameters. The next
// available namespace ID is used for the create operation. Note that
// create-ns does not attach the namespace to a controller, the attach-ns
//command is needed.
type CreateNsCmd struct {
	Device    string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	Size      uint64 `kong:"arg,required,short='s',help='size of ns'"`
	Capacity  uint64 `kong:"arg,required,short='c',help='capacity of ns'"`
	Format    uint8  `kong:"arg,optional,short='f',default='127',help='FLBA size'"`
	Dps       uint8  `kong:"arg,optional,short='d',default='0',help='data protection capabilities'"`
	Multipath uint8  `kong:"arg,optional,short='m',default='0',help='multipath and sharing capabilities'"`
	Anagrpid  uint32 `kong:"arg,optional,short='a',default='0',help='ANA Group Identifier'"`
	Nvmesetid uint16 `kong:"arg,optional,short='i',default='0',help='NVM Set Identifier'"`
	Blocksize uint64 `kong:"arg,optional,short='b',default='0',help='target block size'"`
	Timeout   uint32 `kong:"arg,optional,short='t',default='0',help='timeout value, in milliseconds'"`

	dev *nvme.Device
}

// AfterApply will be called after assignment; used to validate input params
func (cmd *CreateNsCmd) AfterApply() error {
	if cmd.Format != 0xff && cmd.Blocksize != 0 {
		return fmt.Errorf("Invalid specification of both FLBAs and Block Size, please specify only one")
	}

	if cmd.Blocksize != 0 {
		if (cmd.Blocksize & (^cmd.Blocksize + 1)) != cmd.Blocksize {
			return fmt.Errorf("Invalid value for block size %d. Block size must be a power of two", cmd.Blocksize)
		}

		// TODO: Identify Namespace (NSID = ALL), find the flbas value that has
		//       matches the block size (and metadata size == 0)
		// for i, lbaf range ns.lbafs {
		//	if (1 << lbaf.ds) == cmd.Blocksize && lbaf.ms == 0 {
		//		cmd.Format = i
		//      break
		//	}
		//}
		return fmt.Errorf("Block Size option not supported yet")
	}

	if cmd.Format == 0xff {
		return fmt.Errorf("FLBAs corresponding to block size %d not found", cmd.Blocksize)
	}

	return nil
}

// Run will run the Create Namespace Command
func (cmd *CreateNsCmd) Run() error {
	dev := cmd.dev
	if dev == nil {
		var err error
		dev, err = nvme.Open(cmd.Device)
		if err != nil {
			return err
		}
	}

	defer dev.Close()

	nsid, err := dev.CreateNamespace(cmd.Size, cmd.Capacity, cmd.Format, cmd.Dps, cmd.Multipath, cmd.Anagrpid, cmd.Nvmesetid, cmd.Timeout)
	if err != nil {
		return err
	}

	fmt.Printf("Success, created nsid: %d\n", nsid)

	return nil
}
