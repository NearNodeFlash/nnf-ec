package openapi

import (
	"encoding/json"
	"testing"
)

func Test(t *testing.T) {

	type Root struct {
		Oem map[string]interface{} `json:"Oem,omitempty"`
	}

	type OemNested struct {
		Int int
	}

	type Oem struct {
		Bool   bool
		String string
		Slice  []string
		Struct OemNested
	}

	oem := Oem{
		Bool:   true,
		String: "string",
		Slice:  make([]string, 2),
		Struct: OemNested{
			Int: 42,
		},
	}

	oem.Slice[0] = "test0"
	oem.Slice[1] = "test1"

	root := Root{}
	root.Oem = MarshalOem(oem)

	m, err := json.Marshal(root)
	if err != nil {
		t.Error(err)
	}

	root2 := Root{}
	err = json.Unmarshal(m, &root2)
	if err != nil {
		t.Error(err)
	}

	oem2 := Oem{}
	err = UnmarshalOem(root2.Oem, &oem2)
	if err != nil {
		t.Error(err)
	}

	if oem.Bool != oem2.Bool ||
		oem.String != oem2.String ||
		oem.Struct.Int != oem.Struct.Int {
		t.Errorf("Failed marshal/unmarshal of OEM data")
	}

	for i, v := range oem.Slice {
		if v != oem2.Slice[i] {
			t.Errorf("Slice index %d mismatch: Expected: '%s' Actual: '%s'", i, v, oem2.Slice[i])
		}
	}
}
