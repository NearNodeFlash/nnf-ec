package fabric

import (
	"fmt"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
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
	Switch0  = "0"
	Switch1  = "1"
)

func isFabric(fabricId string) bool { return fabricId == FabricId }
func isSwitch(switchId string) bool { return switchId == Switch0 || switchId == Switch1 }

// Get -
func Get(model *sf.FabricCollectionFabricCollection) error {
	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/Fabrics/%s", FabricId)

	return nil
}

// FabricIdGet -
func FabricIdGet(fabricId string, model *sf.FabricV120Fabric) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.FabricType = sf.PC_IE_PP
	model.Switches.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches", fabricId)
	model.Connections.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Connections", fabricId)
	model.Endpoints.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Endpoints", fabricId)

	return nil
}

// FabricIdSwitchesGet -
func FabricIdSwitchesGet(fabricId string, model *sf.SwitchCollectionSwitchCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = 2
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%s", fabricId, Switch0)
	model.Members[1].OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%s", fabricId, Switch1)

	return nil
}

// FabricIdSwitchesSwitchIdGet -
func FabricIdSwitchesSwitchIdGet(fabricId string, switchId string, model *sf.SwitchV140Switch) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	_, err := findSwitchById(switchId)
	if err != nil {
		model.Status.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		model.Status.State = sf.ENABLED_RST
	}

	model.Ports.OdataId = fmt.Sprintf("/redfish/Fabrics/%s/Switches/%s/Ports", fabricId, switchId)

	return nil
}

// FabricIdSwitchesSwitchIdPortsGet -
func FabricIdSwitchesSwitchIdPortsGet(fabricId string, switchId string, model *sf.PortCollectionPortCollection) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	return nil
}

// FabricIdSwitchesSwitchIdPortsPortIdGet -
func FabricIdSwitchesSwitchIdPortsPortIdGet(fabricId string, switchId string, portId string, model *sf.PortV130Port) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	return nil
}

// FabricIdEndpointsGet -
func FabricIdEndpointsGet(fabricId string, model *sf.EndpointCollectionEndpointCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	return nil
}

// FabricIdEndpointsEndpointIdGet -
func FabricIdEndpointsEndpointIdGet(fabricId string, endpointId string, model *sf.EndpointV150Endpoint) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	return nil
}

// FabricIdConnectionsGet -
func FabricIdConnectionsGet(fabricId string, model *sf.ConnectionCollectionConnectionCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	return nil
}

// FabricIdConnectionsPost -
func FabricIdConnectionsPost(fabricId string, model *sf.ConnectionV100Connection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	return nil
}
