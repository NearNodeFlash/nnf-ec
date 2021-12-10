package nvme

import "github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"

func run(device string, f func(dev *nvme.Device) error) error {
	dev, err := nvme.Open(device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return f(dev)
}
