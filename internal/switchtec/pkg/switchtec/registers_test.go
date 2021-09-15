package switchtec

import (
	"testing"
	"unsafe"
)

func TestRegisterLayout(t *testing.T) {
	// Not sure about byte-packing the registers

	assertEqual := func(a uintptr, b uintptr, t *testing.T) {
		if a != b {
			t.Fatalf("%v ! %v", a, b)
		}
	}

	var regs mrpcRegs
	assertEqual(unsafe.Offsetof(regs.command), 2048, t)
	assertEqual(unsafe.Offsetof(regs.status), 2052, t)
	assertEqual(unsafe.Offsetof(regs.ret), 2056, t)
}
