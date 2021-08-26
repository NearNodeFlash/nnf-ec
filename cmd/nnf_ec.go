package main

import (
	"flag"

	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg"
	ec "stash.us.cray.com/rabsw/nnf-ec/pkg/ec"
)

func main() {

	// I would love for this to read a little better and FORCE usage of the command line flags,
	// not make it optional. EC does not need an options interface, it should do everything
	// when Run() is called (if there is no harm to running flag.Parse() twice?).
	// Similarly, the NNF Controller should hide the flags it needs in the call to New()

	nnfOpts := nnf.BindFlags(flag.CommandLine)
	ecOpts := ec.BindFlags(flag.CommandLine)

	flag.Parse()

	c := nnf.NewController(nnfOpts)

	c.Init(ecOpts)
	c.Run()
}
