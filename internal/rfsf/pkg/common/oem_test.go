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
		Struct OemNested
	}

	oem := Oem{
		Bool:   true,
		String: "string",
		Struct: OemNested{
			Int: 42,
		},
	}

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
}

func TestEmpty(t *testing.T) {

	type Oem struct {
		Value string
		Empty string
	}

	data := map[string]interface{}{"Value": "value"}

	oem := Oem{}
	if err := UnmarshalOem(data, &oem); err != nil {
		t.Error(err)
	}

	if oem.Value != "value" {
		t.Errorf("Failed to unmarshal OEM data")
	}
}
