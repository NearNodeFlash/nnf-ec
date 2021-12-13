package nvme

import (
	"fmt"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

// DeleteNsCmd delete the given namespace by sending a namespace management
// command to the provided device. All controllers should be detached from
// the namespace prior to namespace deletion. A namespace ID becomes inactive
// when that namespace is detached or, if the namespace is not already inactive,
// once deleted.
type DeleteNsCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32 `kong:"arg,required,short='n',help='namespace to delete'"`
}

// Run will run the Delete Namespace Command
func (cmd *DeleteNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.DeleteNamespace(cmd.NamespaceID); err != nil {
		return err
	}

	fmt.Printf("Success, deleted nsid: %d\n", cmd.NamespaceID)

	return nil
}
