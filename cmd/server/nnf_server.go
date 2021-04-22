package main

import (
	"fmt"

	remote "stash.us.cray.com/rabsw/nnf-ec/internal/manager-remote"
)

func main() {

	if err := remote.RunController(); err != nil {
		fmt.Printf("Controller Error: %s", err)
	}

	fmt.Println("Exiting...")
}
