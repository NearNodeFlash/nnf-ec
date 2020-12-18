package main

import (
	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg"
)

func main() {
	ec.Run(nnf.Controller)
}
