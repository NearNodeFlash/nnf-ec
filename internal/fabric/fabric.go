package fabric

import (
	"fmt"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"

	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
)

type swtch struct {
	id   string
	path string
	dev  *switchtec.Device
	s    sf.SwitchV140Switch
}

var (
	switches [2]swtch
)

func (s *swtch) isReady() bool {
	return s.dev != nil
}

func (s *swtch) identify(idx int) error {
	s.path = fmt.Sprintf("/dev/switchtec%d", idx)
	dev, err := switchtec.Open(s.path)
	if err != nil {
		return err
	}

	id, err := dev.Identify()
	if err != nil {
		return err
	}

	s.id = fmt.Sprintf("%d", id)
	return nil
}

func findSwitchById(id string) (*swtch, error) {
	for i, s := range switches {
		if !s.isReady() {
			if err := s.identify(i); err != nil {
				return nil, err
			}
		}

		if s.id == id {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("Could not find switch %s", id)
}

const (
	FabricId = "Rabbit"
)

// Get -
func Get(model *sf.FabricCollectionFabricCollection) error {
	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/Fabrics/%s", FabricId)

	return nil
}

// FabricIdGet -
func FabricIdGet(id string, model *sf.FabricV120Fabric) error {
	model.FabricType = sf.PC_IE_PP
	model.Switches.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches", id)
	model.Connections.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Connections", id)
	model.Endpoints.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Endpoints", id)

	return nil
}

// FabricIdSwitchesGet- 
func FabricIdSwitchesGet(model *sf.SwitchCollectionSwitchCollection) error {
	model.MembersodataCount = 2
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%d", FabricId, 0)
	model.Members[1].OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%d", FabricId, 1)

	return nil
}

// FabricIdSwitchesSwitchIdGet
func FabricIdSwitchesSwitchIdGet(id string, model *sf.SwitchV140Switch) error {
	_, err := findSwitchById(id)
	if err != nil {
		model.Status.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		model.Status.State = sf.ENABLED_RST
	}

	model.Ports.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%s/Ports", FabricId, id)

	return nil
}

