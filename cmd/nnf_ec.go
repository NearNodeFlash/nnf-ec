package main

import (
	"flag"

	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg"
)

func main() {
	mock := flag.Bool("mock", false, "enable mock interfaces and devices")
	serve := flag.Bool("serve", false, "enable self hosted http server")
	flag.Parse()

	ctrl := nnf.NewController(*mock)
	if *serve {
		ctrl.ListenAndServe()
	} else {
		ctrl.Run()
	}
}
