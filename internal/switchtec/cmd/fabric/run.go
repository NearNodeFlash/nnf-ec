package cmd

import (
	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

func run(device string, f func(*switchtec.Device) error) error {
	dev, err := switchtec.Open(device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return f(dev)
}
