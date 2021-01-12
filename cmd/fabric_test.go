package main

import (
	"testing"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric"
)

func TestInit(t *testing.T) {
	if err := fabric.Initialize(); err != nil {
		t.Error(err)
	}
}

// TODO: Get all the *Get endponts and test them
