package fabric

import (
	"testing"
)

func TestFabric(t *testing.T) {
	if err := Initialize(NewMockSwitchtecController()); err != nil {
		t.Error(err)
	}
}

func TestFabricSwitchesDown(t *testing.T) {
	ctrl := NewMockSwitchtecController()
	mock, ok := ctrl.(*MockSwitchtecController)
	if !ok {
		t.Fatal()
	}

	mock.SetSwitchNotExists(0)
	mock.SetSwitchNotExists(1)

	if err := Initialize(ctrl); err != nil {
		t.Error(err)
	}
}
