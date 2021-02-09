/*
 * Near Node Flash
 *
 * This file contains the declaration of the Near-Node Flash
 * Element Controller.
 *
 * Author: Nate Roiger
 *
 * Copyright 2020 Hewlett Packard Enterprise Development LP
 */

package nnf

import (
	ec "stash.us.cray.com/rabsw/nnf-ec/ec"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric-manager"
	"stash.us.cray.com/rabsw/nnf-ec/internal/nvme-namespace-manager"
)

var (
	// Controller - the Near-Node Flash Element Controller
	Controller = &ec.Controller{
		Name:     "Near Node Flash",
		Port:     "50057",
		Version:  "v2",
		Routers:  NewDefaultApiRouters(fabric.NewSwitchtecController()),
	}
)

// NewDefaultApiRouters -
func NewDefaultApiRouters(switchCtrl fabric.SwitchtecControllerInterface) ec.Routers {

	routers := make([]ec.Router, 2)
	
	routers[0] = fabric.NewDefaultApiRouter(
		fabric.NewDefaultApiService(),
		switchCtrl,
	)

	routers[1] = nvmenamespace.NewDefaultApiRouter(
		nvmenamespace.NewDefaultApiService(),
	)

	return routers
}