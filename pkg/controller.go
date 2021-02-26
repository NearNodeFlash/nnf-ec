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
	ec "stash.us.cray.com/rabsw/ec"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric-manager"
	nvmenamespace "stash.us.cray.com/rabsw/nnf-ec/internal/nvme-namespace-manager"
)

// NewController - Create a new NNF Element Controller with the desired mocking behavior
func NewController(mock bool) *ec.Controller {
	switchCtrl := fabric.NewSwitchtecController()
	nvmeCtrl := nvmenamespace.NewNvmeController()

	if mock {
		switchCtrl = fabric.NewMockSwitchtecController()
		nvmeCtrl = nvmenamespace.NewMockNvmeController()
	}

	return &ec.Controller{
		Name:    "Near Node Flash",
		Port:    50057,
		Version: "v2",
		Routers: NewDefaultApiRouters(switchCtrl, nvmeCtrl),
	}
}

// NewDefaultApiRouters -
func NewDefaultApiRouters(
	switchCtrl fabric.SwitchtecControllerInterface,
	nvmeCtrl nvmenamespace.NvmeControllerInterface) ec.Routers {

	routers := make([]ec.Router, 2)

	routers[0] = fabric.NewDefaultApiRouter(
		fabric.NewDefaultApiService(),
		switchCtrl,
	)

	routers[1] = nvmenamespace.NewDefaultApiRouter(
		nvmenamespace.NewDefaultApiService(),
		nvmeCtrl,
	)

	return routers
}
