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
	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric"
)

var (
	// Controller - the Near-Node Flash Element Controller
	Controller = &ec.Controller{
		Name:     "Near Node Flash",
		Port:     "50057",
		Version:  "v2",
		Router:   NewDefaultApiRouter(NewDefaultApiService()),
		InitFunc: ControllerInitialize,
	}
)

// ControllerInitialize -
func ControllerInitialize() error {
	return fabric.Initialize()
}
