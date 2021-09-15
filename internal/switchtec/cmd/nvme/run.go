package nvme

import "stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/nvme"

func run(device string, f func(dev *nvme.Device) error) error {
	dev, err := nvme.Open(device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return f(dev)
}
