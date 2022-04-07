package nvme

import (
	"fmt"
	"strings"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

type DiscoverCmd struct {
	Regexp string `kong:"arg,required,help='regular expression used to filter drives by model number, serial number, or NQN'"`
}

func (cmd *DiscoverCmd) Run() error {
	devices, err := nvme.DeviceList(cmd.Regexp)
	if err != nil {
		return err
	}

	fmt.Println("Device List:")
	fmt.Println(strings.Join(devices, "\n"))

	return nil
}
