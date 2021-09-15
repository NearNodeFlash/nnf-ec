package nvme

import (
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/nvme"
)

// ListNsCmd For the specified controller handle, show the namespace list
// in the associated NVMe subsystem, optionally starting with a given nsid.
type ListNsCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32 `kong:"optional,default='1',help='First NSID returned list should start from.'"`
	All         bool   `kong:"optional,default='false',help='Show all namespaces in the subsystem, whether attached or inactive.'"`
}

// Run will run the List Namespace Command.
func (cmd *ListNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	list, err := dev.IdentifyNamespaceList(cmd.NamespaceID-1, cmd.All)
	if err != nil {
		return err
	}

	for idx, id := range list {
		if id != 0 {
			fmt.Printf("[%4d]:%#x\n", idx, id)
		}
	}

	return nil
}
