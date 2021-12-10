package nvme

import (
	"fmt"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

// AttachNsCmd attaches the given namespace to the given controller or comma-sep list of
// controllers. ID of the given namespace becomes active upon attachment to
// a controller. A namespace must be attached to a controller before IO
// commands may be directed to that namespace.
type AttachNsCmd struct {
	Device      string   `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32   `kong:"arg,required,short='n',help='namespace to attach'"`
	Controllers []uint16 `kong:"arg,required,short='c',help='comma-sep controller list'"`
}

// Run will run the Attach Namespace Command
func (cmd *AttachNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.AttachNamespace(cmd.NamespaceID, cmd.Controllers); err != nil {
		return err
	}

	fmt.Printf("Success, attached nsid: %d\n", cmd.NamespaceID)

	return nil
}
