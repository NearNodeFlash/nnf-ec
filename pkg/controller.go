package nnf

import ec "stash.us.cray.com/rabsw/nnf-ec/ec"

var (
	// Controller for Near-Node Flash Element Controller
	Controller = &ec.Controller{
		Name:     "Near Node Flash",
		Port:     "50053",
		Version:  "v1",
		Servicer: NewDefaultApiService(),
	}
)
