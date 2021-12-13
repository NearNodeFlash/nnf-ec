package main

import (
	"fmt"

	remote "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-remote"
)

func main() {

	if err := remote.RunController(); err != nil {
		fmt.Printf("Controller Error: %s", err)
	}

	fmt.Println("Exiting...")
}
