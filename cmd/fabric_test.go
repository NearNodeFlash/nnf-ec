package main

import (
	"testing"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric"
)

func TestFabric(t *testing.T) {
	if err := fabric.Initialize(fabric.NewMockSwitchtecController()); err != nil {
		t.Error(err)
	}
}

// TODO: Get all the *Get endponts and test them
