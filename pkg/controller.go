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
	"flag"

	ec "stash.us.cray.com/rabsw/ec"

	fabric "stash.us.cray.com/rabsw/nnf-ec/internal/manager-fabric"
	nnf "stash.us.cray.com/rabsw/nnf-ec/internal/manager-nnf"
	nvme "stash.us.cray.com/rabsw/nnf-ec/internal/manager-nvme"
)

type options struct {
	mock  bool // Enable mock interfaces for Switches, NVMe, and NNF
	cli   bool // Enable CLI commands instead of binary
}

func newDefaultOptions() *options {
	return &options{mock: false, cli: false}
}

func BindFlags(fs *flag.FlagSet) *options {
	opts := newDefaultOptions()

	fs.BoolVar(&opts.mock,  "mock",  opts.mock,  "Enable mock (simulated) environment.")
	fs.BoolVar(&opts.cli,   "cli",   opts.cli,   "Enable CLI interfaces with devices, instead of raw binary.")
	
	nvme.BindFlags(fs)

	return opts
}

// NewController - Create a new NNF Element Controller with the desired mocking behavior
func NewController(opts* options) *ec.Controller {
	if opts == nil {
		opts = newDefaultOptions()
	}

	switchCtrl := fabric.NewSwitchtecController()
	nvmeCtrl := nvme.NewSwitchtecNvmeController()
	nnfCtrl := nnf.NewNnfController()

	if opts.cli {
		switchCtrl = fabric.NewSwitchtecCliController()
		nvmeCtrl = nvme.NewCliNvmeController()
	}

	if opts.mock {
		switchCtrl = fabric.NewMockSwitchtecController()
		nvmeCtrl = nvme.NewMockNvmeController()
		nnfCtrl = nnf.NewMockNnfController()
	}

	return &ec.Controller{
		Name:    "Near Node Flash",
		Port:    50057,
		Version: "v2",
		Routers: NewDefaultApiRouters(switchCtrl, nvmeCtrl, nnfCtrl),
	}
}

// NewDefaultApiRouters -
func NewDefaultApiRouters(switchCtrl fabric.SwitchtecControllerInterface, nvmeCtrl nvme.NvmeController, nnfCtrl nnf.NnfControllerInterface) ec.Routers {

	routers := []ec.Router{
		fabric.NewDefaultApiRouter(fabric.NewDefaultApiService(), switchCtrl),
		nvme.NewDefaultApiRouter(nvme.NewDefaultApiService(), nvmeCtrl),
		nnf.NewDefaultApiRouter(nnf.NewDefaultApiService(), nnfCtrl),
	}

	return routers
}
