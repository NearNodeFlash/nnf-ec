package nvme

import (
	"fmt"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

// DetachNsCmd detaches the given namespace from the given controller;
// de-activates the given namespace's ID. A namespace must be attached
// to a controller before IO commands may be directed to that namespace.
type DetachNsCmd struct {
	Device      string   `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32   `kong:"arg,required,short='n',help='namespace to attach'"`
	Controllers []uint16 `kong:"arg,required,short='c',help='comma-sep controller list'"`
}

// Run will run the Detach Namespace Command
func (cmd *DetachNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.DetachNamespace(cmd.NamespaceID, cmd.Controllers); err != nil {
		return err
	}

	fmt.Printf("Success, detached nsid: %d\n", cmd.NamespaceID)

	return nil
}
