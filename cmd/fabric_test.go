package main

import (
	"testing"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric-manager"
)

func TestFabric(t *testing.T) {

	if err := fabric.Initialize(fabric.NewMockSwitchtecController()); err != nil {
		t.Error(err)
	}
}

func TestFabricSwitchesDown(t *testing.T) {
	ctrl := fabric.NewMockSwitchtecController()
	mock, ok := ctrl.(*fabric.MockSwitchtecController)
	if !ok {
		t.Fatal()
	}

	mock.SetSwitchNotExists(0)
	mock.SetSwitchNotExists(1)

	if err := fabric.Initialize(ctrl); err != nil {
		t.Error(err)
	}
}

// TODO: Get all the *Get endponts and test them
