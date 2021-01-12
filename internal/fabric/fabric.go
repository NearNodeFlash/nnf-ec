package fabric

import (
	"fmt"
	"strconv"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
)

const (
	FabricId = "Rabbit"
)

type Fabric struct {
	id             string
	switches       []Switch
	endpoints      []Endpoint
	endpointGroups []EndpointGroup
	connections    []Connection

	managmentPortCount  int
	upstreamPortCount   int
	downstreamPortCount int

	managementEndpointCount int
	upstreamEndpointCount   int
	downstreamEndpointCount int
}

type Switch struct {
	id     string
	paxId  int32
	path   string
	dev    *switchtec.Device
	config *SwitchConfig
	ports  []Port

	managementPortCount int
	upstreamPortCount   int
	downstreamPortCount int
}

type Port struct {
	id     string
	swtch  *Switch
	config *PortConfig

	portType sf.PortV130PortType
}

type Endpoint struct {
	id    string
	ports []*Port
}

type EndpointGroup struct {
	id        string
	endpoints []*Endpoint
	connection *Connection
}

type Connection struct {
	endpointGroup *EndpointGroup
}

var fabric = &Fabric{
	id: FabricId,
}

func isFabric(fabricId string) bool     { return fabricId == FabricId }
func isSwitch(switchId string) bool     { _, err := fabric.findSwitch(switchId); return err == nil }
func isEndpoint(endpointId string) bool { _, err := fabric.findEndpoint(endpointId); return err == nil }
func isEndpointGroup(groupId string) bool {
	_, err := fabric.findEndpointGroup(groupId)
	return err == nil
}

func (f *Fabric) findSwitch(switchId string) (*Switch, error) {
	for _, s := range f.switches {
		if !s.isReady() {
			if err := s.identify(); err != nil {
				return nil, err
			}
		}

		if s.id == switchId {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("Could not find switch %s", switchId)
}

func (f *Fabric) findSwitchPort(switchId string, portId string) (*Port, error) {
	s, err := f.findSwitch(switchId)
	if err != nil {
		return nil, err
	}

	for _, p := range s.ports {
		if p.id == portId {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("Could not find switch %s port %s", switchId, portId)
}

// findPort - Find's the i'th port of portType in the fabric
func (f *Fabric) findPort(portType sf.PortV130PortType, idx int) *Port {
	switch portType {
	case sf.MANAGEMENT_PORT_PV130PT:
		return f.switches[idx].findPort(portType, 0)
	case sf.UPSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.upstreamPortCount {
				return s.findPort(portType, idx)
			}
			idx = idx - s.upstreamPortCount
		}
	case sf.DOWNSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.downstreamPortCount {
				return s.findPort(portType, idx)
			}
			idx = idx - s.downstreamPortCount
		}
	}

	return nil
}

func (f *Fabric) isManagementEndpoint(endpointIndex int) bool {
	return endpointIndex == 0
}

func (f *Fabric) isUpstreamEndpoint(idx int) bool {
	return !f.isManagementEndpoint(idx) && idx-f.managementEndpointCount < f.upstreamEndpointCount
}

func (f *Fabric) isDownstreamEndpoint(idx int) bool {
	return !f.isUpstreamEndpoint(idx) && idx-f.managementEndpointCount-f.upstreamEndpointCount < f.downstreamEndpointCount
}

func (f *Fabric) getUpstreamEndpointRelativePortIndex(idx int) int {
	return idx - f.managementEndpointCount
}

func (f *Fabric) getDownstreamEndpointRelativePortIndex(idx int) int {
	return (idx - (f.managementEndpointCount + f.upstreamEndpointCount)) / (f.managementEndpointCount + f.upstreamEndpointCount)
}

func (f *Fabric) findEndpoint(endpointId string) (*Endpoint, error) {
	id, err := strconv.Atoi(endpointId)
	if err != nil {
		return nil, ec.ErrNotFound
	}

	if !(id < len(f.endpoints)) {
		return nil, ec.ErrNotFound
	}

	return &f.endpoints[id], nil
}

func (f *Fabric) findEndpointGroup(endpointGroupId string) (*EndpointGroup, error) {
	id, err := strconv.Atoi(endpointGroupId)
	if err != nil {
		return nil, ec.ErrNotFound
	}

	if !(id < len(f.endpointGroups)) {
		return nil, ec.ErrNotFound
	}

	return &f.endpointGroups[id], nil
}

func (s *Switch) isReady() bool {
	return s.dev != nil
}

func (s *Switch) identify() error {

	for i := 0; i < len(fabric.switches); i++ {

		path := fmt.Sprintf("/dev/switchtec%d", i)

		dev, err := switchtec.Open(path)
		if err != nil {
			return err
		}

		paxId, err := dev.Identify()
		if err != nil {
			return err
		}

		if id := strconv.Itoa(int(paxId)); id == s.id {
			s.dev = dev
			s.path = path
			return nil
		}

		dev.Close()
	}

	return fmt.Errorf("Could not identify switch %s", s.id)
}

// findPort - Finds the i'th port of portType in the switch
func (s *Switch) findPort(portType sf.PortV130PortType, idx int) *Port {
	for _, p := range s.ports {
		if p.portType == portType {
			if idx == 0 {
				return &p
			}
			idx--
		}
	}

	panic(fmt.Sprintf("Switch Port %d Not Found", idx))
}

// Initialize
func Initialize() error {
	f := fabric

	if err := loadConfig(); err != nil {
		return err
	}

	fabric.switches = make([]Switch, len(Config.Switches))
	for switchIdx, switchConf := range Config.Switches {
		s := Switch{
			config: &switchConf,
			ports:  make([]Port, len(switchConf.Ports)),
		}

		for portIdx, portConf := range switchConf.Ports {
			portType := portConf.getPortType()
			switch portType {
			case sf.MANAGEMENT_PORT_PV130PT:
				s.managementPortCount++
			case sf.UPSTREAM_PORT_PV130PT:
				s.upstreamPortCount++
			case sf.DOWNSTREAM_PORT_PV130PT:
				s.downstreamPortCount++
			}

			s.ports[portIdx] = Port{
				swtch:    &s, // I dont think this is needed
				config:   &portConf,
				portType: portType,
			}
		}

		fabric.switches[switchIdx] = s
	}

	// Tally the switch ports and apply to the fabric
	for _, s := range fabric.switches {
		fabric.managmentPortCount += s.managementPortCount
		fabric.upstreamPortCount += s.upstreamPortCount
		fabric.downstreamPortCount += s.downstreamPortCount
	}

	// configure the endpoints

	fabric.managementEndpointCount = 1
	fabric.upstreamEndpointCount = fabric.upstreamPortCount
	fabric.downstreamEndpointCount = (fabric.managementEndpointCount + fabric.upstreamEndpointCount) * fabric.downstreamPortCount

	fabric.endpoints = make([]Endpoint, 1+fabric.managementEndpointCount+fabric.upstreamEndpointCount+fabric.downstreamEndpointCount)
	for endpointIdx, endpoint := range fabric.endpoints {

		if fabric.isManagementEndpoint(endpointIdx) {
			endpoint.ports = make([]*Port, len(fabric.switches))
			for switchIdx, s := range fabric.switches {
				endpoint.ports[switchIdx] = s.findPort(sf.MANAGEMENT_PORT_PV130PT, 0)
			}
		} else if fabric.isUpstreamEndpoint(endpointIdx) {
			endpoint.ports = make([]*Port, 1)
			endpoint.ports[0] = fabric.findPort(sf.UPSTREAM_PORT_PV130PT, fabric.getUpstreamEndpointRelativePortIndex(endpointIdx))
		} else if fabric.isDownstreamEndpoint(endpointIdx) {
			endpoint.ports = make([]*Port, 1)
			endpoint.ports[0] = fabric.findPort(sf.DOWNSTREAM_PORT_PV130PT, fabric.getDownstreamEndpointRelativePortIndex(endpointIdx))
		}
	}

	// configure the endpoint groups

	mangementAndUpstreamEndpointCount := fabric.managementEndpointCount + fabric.upstreamEndpointCount
	fabric.endpointGroups = make([]EndpointGroup, mangementAndUpstreamEndpointCount)
	for endpointGroupIdx, endpointGroup := range fabric.endpointGroups {

		endpointGroup.endpoints = make([]*Endpoint, 1+f.downstreamEndpointCount)
		endpointGroup.endpoints[0] = &f.endpoints[endpointGroupIdx] // Mgmt or USP

		for idx, _ := range endpointGroup.endpoints[1:] {
			endpointGroup.endpoints[1+idx] = &f.endpoints[endpointGroupIdx+mangementAndUpstreamEndpointCount+idx*(mangementAndUpstreamEndpointCount)]
		}
	}

	// configure the connections

	fabric.connections = make([]Connection, mangementAndUpstreamEndpointCount)
	for connectionIdx, connection := range fabric.connections {
		epg := &fabric.endpointGroups[connectionIdx]
		epg.connection = &connection
		connection.endpointGroup = epg
	}

	return nil
}

// Get -
func Get(model *sf.FabricCollectionFabricCollection) error {
	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s", fabric.id)

	return nil
}

// FabricIdGet -
func FabricIdGet(fabricId string, model *sf.FabricV120Fabric) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.FabricType = sf.PC_IE_PP
	model.Switches.OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches", fabricId)
	model.Connections.OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections", fabricId)
	model.Endpoints.OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints", fabricId)
	model.EndpointGroups.OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/EndpointGroups", fabricId)

	return nil
}

// FabricIdSwitchesGet -
func FabricIdSwitchesGet(fabricId string, model *sf.SwitchCollectionSwitchCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.switches))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, s := range fabric.switches {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s", fabricId, s.id)
	}

	return nil
}

// FabricIdSwitchesSwitchIdGet -
func FabricIdSwitchesSwitchIdGet(fabricId string, switchId string, model *sf.SwitchV140Switch) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	model.Id = switchId
	model.Status.Health = sf.OK_RH
	model.SwitchType = sf.PC_IE_PP

	_, err := fabric.findSwitch(switchId)
	if err != nil {
		model.Status.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		model.Status.State = sf.ENABLED_RST
	}

	// TODO: None of this is present in the switchtec-user util - how to get this information?
	model.FirmwareVersion = "TODO"
	model.Model = "TODO"
	model.Manufacturer = "TODO"
	model.PartNumber = "TODO"
	model.SerialNumber = "TODO"

	model.Ports.OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s/Ports", fabricId, switchId)

	return nil
}

// FabricIdSwitchesSwitchIdPortsGet -
func FabricIdSwitchesSwitchIdPortsGet(fabricId string, switchId string, model *sf.PortCollectionPortCollection) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	s, err := fabric.findSwitch(switchId)
	if err != nil {
		return err
	}

	model.MembersodataCount = int64(len(s.ports))
	model.Members = make([]sf.OdataV4IdRef, len(s.ports))
	for idx, port := range s.ports {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s/Ports/%s", fabricId, switchId, port.id)
	}

	return nil
}

// FabricIdSwitchesSwitchIdPortsPortIdGet -
func FabricIdSwitchesSwitchIdPortsPortIdGet(fabricId string, switchId string, portId string, model *sf.PortV130Port) error {
	if !isFabric(fabricId) || !isSwitch(switchId) {
		return ec.ErrNotFound
	}

	p, err := fabric.findSwitchPort(switchId, portId)
	if err != nil {
		return err
	}

	model.Name = p.config.Name
	model.Id = p.id

	model.PortProtocol = sf.PC_IE_PP
	model.PortMedium = sf.ELECTRICAL_PV130PM
	model.PortType = p.portType
	model.PortId = strconv.Itoa(p.config.Port)

	model.Width = int64(p.config.Width)
	model.ActiveWidth = 0 // TODO

	//model.MaxSpeedGbps = 0 // TODO
	//model.CurrentSpeedGbps = 0 // TODO

	model.LinkState = sf.ENABLED_PV130LST
	//model.LinkStatus = // TODO

	//model.Links.AssociatedEndpoints // TODO

	return nil
}

// FabricIdEndpointsGet -
func FabricIdEndpointsGet(fabricId string, model *sf.EndpointCollectionEndpointCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.endpoints))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, _ := range fabric.endpoints {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%d", fabricId, idx)
	}

	return nil
}

// FabricIdEndpointsEndpointIdGet -
func FabricIdEndpointsEndpointIdGet(fabricId string, endpointId string, model *sf.EndpointV150Endpoint) error {
	if !isFabric(fabricId) || !isEndpoint(endpointId) {
		return ec.ErrNotFound
	}

	ep, err := fabric.findEndpoint(endpointId)
	if err != nil {
		return err
	}

	model.EndpointProtocol = sf.PC_IE_PP
	model.PciId = sf.EndpointV150PciId{
		ClassCode:      "", // TODO
		DeviceId:       "", // TODO
		FunctionNumber: 0,  // TODO
	}

	model.Links.Ports = make([]sf.OdataV4IdRef, len(ep.ports))
	for idx, port := range ep.ports {
		model.Links.Ports[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s/Ports/%s", fabricId, port.swtch.id, port.id)
	}

	return nil
}

func FabricIdEndpointGroupsGet(fabricId string, model *sf.EndpointGroupCollectionEndpointGroupCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.endpointGroups))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, _ := range fabric.endpointGroups {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/EndpointGroups/%d", fabricId, idx)
	}

	return nil
}

func FabricIdEndpointGroupsEndpointIdGet(fabricId string, groupId string, model *sf.EndpointGroupV130EndpointGroup) error {
	if !isFabric(fabricId) || !isEndpointGroup(groupId) {
		return ec.ErrNotFound
	}

	epg, err := fabric.findEndpointGroup(groupId)
	if err != nil {
		return err
	}

	model.Links.EndpointsodataCount = int64(len(epg.endpoints))
	model.Links.Endpoints = make([]sf.OdataV4IdRef, model.Links.EndpointsodataCount)
	for idx, ep := range epg.endpoints {
		model.Links.Endpoints[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, ep.id)
	}

	model.Links.ConnectionsodataCount = int64(len(epg.connections))
	model.Links.Connections = make([]sf.OdataV4IdRef, model.Links.ConnectionsodataCount) // TODO
	for idx, c := range epg.connections {
		model.Links.Connections[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections/%s", c.id)
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
