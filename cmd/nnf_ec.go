package main

import (
	"flag"

	"stash.us.cray.com/rabsw/ec"
	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg"
)

func main() {
	
	mock := flag.Bool("mock", false, "enable mock interfaces and devices")
	opts := ec.BindFlags(flag.CommandLine)
	flag.Parse()

	c := nnf.NewController(*mock)

	c.Run(opts)
}
