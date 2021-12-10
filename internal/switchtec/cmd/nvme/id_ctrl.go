package nvme

import (
	"fmt"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

// IdCtrlCmd shows secondary controller list associated with the
// primary controller of the given device.
type IdCtrlCmd struct {
	Device string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
}

// Run will run the Identify Controller Command.
func (cmd *IdCtrlCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	ctrl, err := dev.IdentifyController()
	if err != nil {
		return err
	}

	fmt.Printf("Identify Controller:")
	fmt.Printf("%+v", ctrl)

	return nil
}

type IdNamespaceCtrls struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceId uint32 `kong:"arg,required,short='n',help='Namespace to identify.'"`
}

func (cmd *IdNamespaceCtrls) Run() error {
	return run(cmd.Device, func(dev *nvme.Device) error {
		ctrlList, err := dev.IdentifyNamespaceControllerList(cmd.NamespaceId)
		if err != nil {
			return err
		}

		fmt.Printf("Controller List for Namespace ID %d:\n", cmd.NamespaceId)
		fmt.Printf("%+v", ctrlList)

		return nil
	})
}
