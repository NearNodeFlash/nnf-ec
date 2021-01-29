package fabric

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	openapi "stash.us.cray.com/sp/rfsf-openapi/pkg/common"
	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

const (
	FabricId = "Rabbit"
)

type Fabric struct {
	ctrl SwitchtecControllerInterface

	id     string
	config *ConfigFile

	switches       []Switch
	endpoints      []Endpoint
	endpointGroups []EndpointGroup
	connections    []Connection

	managementEndpointCount int
	upstreamEndpointCount   int
	downstreamEndpointCount int
}

type Switch struct {
	id     string
	paxId  int32
	path   string
	dev    SwitchtecDeviceInterface
	config *SwitchConfig
	ports  []Port

	fabric   *Fabric
	mgmtPort *Port

	// Information is cached on switch initialization
	model           string
	manufacturer    string
	serialNumber    string
	firmwareVersion string
}

type Port struct {
	id     string
	idx    int
	typ    sf.PortV130PortType
	swtch  *Switch
	config *PortConfig

	endpoints []*Endpoint
}

type Endpoint struct {
	id           string
	endpointType sf.EndpointV150EntityType
	ports        []*Port

	pdfid         uint16
	bound         bool
	boundPaxId    uint8
	boundHvdPhyId uint8
	boundHvdLogId uint8
}

type EndpointGroup struct {
	id         string
	endpoints  []*Endpoint
	connection *Connection
}

type Connection struct {
	endpointGroup *EndpointGroup
}

var fabric Fabric

func isFabric(id string) bool        { return id == FabricId }
func isSwitch(id string) bool        { _, err := fabric.findSwitch(id); return err == nil }
func isEndpoint(id string) bool      { _, err := fabric.findEndpoint(id); return err == nil }
func isEndpointGroup(id string) bool { _, err := fabric.findEndpointGroup(id); return err == nil }
func isConnection(id string) bool    { _, err := fabric.findConnection(id); return err == nil }

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

// findPort - Finds the i'th port of portType in the fabric
func (f *Fabric) findPort(portType sf.PortV130PortType, idx int) *Port {
	switch portType {
	case sf.MANAGEMENT_PORT_PV130PT:
		return f.switches[idx].findPort(portType, 0)
	case sf.UPSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.config.UpstreamPortCount {
				return s.findPort(portType, idx)
			}
			idx = idx - s.config.UpstreamPortCount
		}
	case sf.DOWNSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.config.DownstreamPortCount {
				return s.findPort(portType, idx)
			}
			idx = idx - s.config.DownstreamPortCount
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
	return idx >= (f.managementEndpointCount + f.upstreamEndpointCount)
}

func (f *Fabric) getUpstreamEndpointRelativePortIndex(idx int) int {
	return idx - f.managementEndpointCount
}

func (f *Fabric) getDownstreamEndpointRelativePortIndex(idx int) int {
	return (idx - (f.managementEndpointCount + f.upstreamEndpointCount)) / (1 /*PF*/ + f.managementEndpointCount + f.upstreamEndpointCount)
}

func (f *Fabric) getDownstreamEndpointIndex(deviceIdx int, functionIdx int) int {
	return (deviceIdx * (1 /*PF*/ + f.managementEndpointCount + f.upstreamEndpointCount)) + functionIdx
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

func (f *Fabric) findConnection(connectionId string) (*Connection, error) {
	id, err := strconv.Atoi(connectionId)
	if err != nil {
		return nil, ec.ErrNotFound
	}

	if !(id < len(f.connections)) {
		return nil, ec.ErrNotFound
	}

	return &f.connections[id], nil
}

func (s *Switch) isReady() bool {
	return s.dev != nil
}

func (s *Switch) identify() error {
	f := s.fabric
	for i := 0; i < len(f.switches); i++ {

		path := fmt.Sprintf("/dev/switchtec%d", i)

		log.Debugf("Identify Switch %s: Opening %s", s.id, path)
		dev, err := f.ctrl.Open(path)
		if os.IsNotExist(err) {
			log.WithError(err).Debugf("path %s", path)
			continue
		} else if err != nil {
			log.WithError(err).Warnf("Identify Switch %s: Open Error", s.id)
			return err
		}

		paxId, err := dev.Identify()
		if err != nil {
			log.WithError(err).Warnf("Identify Switch %s: Identify Error", s.id)
			return err
		}

		log.Infof("Identify Switch %s: Device ID: %d", s.id, paxId)
		if id := strconv.Itoa(int(paxId)); id == s.id {
			s.dev = dev
			s.path = path
			s.paxId = paxId

			log.Infof("Identify Switch %s: Loading Mfg Info", s.id)

			s.model = s.getModel()
			s.manufacturer = s.getManufacturer()
			s.serialNumber = s.getSerialNumber()
			s.firmwareVersion = s.getFirmwareVersion()

			return nil
		}

		dev.Close()
	}

	return fmt.Errorf("Identify Switch %s: Could Not ID Switch", s.id) // TODO: Switch not found
}

func (s *Switch) getStatus() (stat sf.ResourceStatus) {

	if s.dev == nil {
		stat.Health = sf.CRITICAL_RH
		stat.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		stat.Health = sf.OK_RH
		stat.State = sf.ENABLED_RST
	}

	return stat
}

func (s *Switch) getDeviceStringByFunc(f func(dev SwitchtecDeviceInterface) (string, error)) string {
	if s.dev != nil {
		ret, err := f(s.dev)
		if err != nil {
			log.WithError(err).Warnf("Failed to retrieve device string")
		}

		return ret
	}

	return ""
}

func (s *Switch) getModel() string {
	return s.getDeviceStringByFunc(func(dev SwitchtecDeviceInterface) (string, error) {
		return dev.GetModel()
	})
}

func (s *Switch) getManufacturer() string {
	return s.getDeviceStringByFunc(func(dev SwitchtecDeviceInterface) (string, error) {
		return dev.GetManufacturer()
	})
}

func (s *Switch) getSerialNumber() string {
	return s.getDeviceStringByFunc(func(dev SwitchtecDeviceInterface) (string, error) {
		return dev.GetSerialNumber()
	})
}

func (s *Switch) getFirmwareVersion() string {
	return s.getDeviceStringByFunc(func(dev SwitchtecDeviceInterface) (string, error) {
		return dev.GetFirmwareVersion()
	})
}

// findPort - Finds the i'th port of portType in the switch
func (s *Switch) findPort(portType sf.PortV130PortType, idx int) *Port {
	for _, p := range s.ports {
		if p.typ == portType {
			if idx == 0 {
				return &p
			}
			idx--
		}
	}

	panic(fmt.Sprintf("Switch Port %d Not Found", idx))
}

func (s *Switch) isDown() bool {
	return !s.isReady()
}

func (p *Port) LinkStatus() error {
	// TODO
	return nil
}

func (p *Port) Initialize() error {

	if p.swtch.isDown() {
		log.Warnf("port %d switch is down", p.id)
		return nil
	}

	if err := p.LinkStatus(); err != nil {
		log.WithError(err).Warnf("Failed to read port %d link status", p.id)
	}

	/*
		switch p.typ {
		case sf.DOWNSTREAM_PORT_PV130PT:

			processPort := func(port *Port) func(*switchtec.DumpEpPortDevice) error {
				return func(epPort *switchtec.DumpEpPortDevice) error {
					if switchtec.EpPortType(epPort.Hdr.Typ) != switchtec.DeviceEpPortType {
						log.Errorf("Port %d is down", p.id)
						// Port & Associated Endpoints are Down/Unreachable
						//p.Down() // TODO
					}

					for idx, f := range epPort.Ep.Functions {

						if int(f.FunctionID) > len(p.endpoints) {
							break
						}

						ep := p.endpoints[idx]

						ep.pdfid = f.PDFID
						ep.bound = f.Bound != 0
						ep.boundPaxId = f.BoundPAXID
						ep.boundHvdPhyId = f.BoundHVDPhyPID
						ep.boundHvdLogId = f.BoundHVDLogPID
					}

					return nil
				}
			}

			log.Infof("Switch %s enumerting endpoint %d", p.swtch.id, p.config.Port)
			p.swtch.dev.EnumerateEndpoint(uint8(p.config.Port), processPort(p))
		}
	*/
	return nil
}

func (c *Connection) Initialize() error {
	/*
		endpointGroup := c.endpointGroup
		initiatorEndpoint := endpointGroup.endpoints[0]

		for idx, downstreamEndpoint := range endpointGroup.endpoints[1:] {

			for _, port := range initiatorEndpoint.ports {

				switch port.typ {
				case sf.MANAGEMENT_PORT_PV130PT:
					if len(downstreamEndpoint.ports) != 1 {
						log.Panicf("Logical Error: Downstream Endpoint %d has multiple ports", downstreamEndpoint.id)
					}

					if downstreamEndpoint.ports[0].swtch == port.swtch {
						if err := port.swtch.dev.Bind(uint8(port.config.Port), uint8(idx), downstreamEndpoint.pdfid); err != nil {
							log.WithError(err).Warnf("Failed to bind port")
						}
					}
				case sf.UPSTREAM_PORT_PV130PT:
					if err := port.swtch.dev.Bind(uint8(port.config.Port), uint8(idx), downstreamEndpoint.pdfid); err != nil {
						log.WithError(err).Warnf("Failed to bind port")
					}
				}
			}
		}
	*/

	return nil
}

// Initialize
func Initialize(ctrl SwitchtecControllerInterface) error {

	fabric = Fabric{
		id:   FabricId,
		ctrl: ctrl,
	}
	f := &fabric

	log.SetLevel(log.DebugLevel)
	log.Infof("Initialize %s Fabric", f.id)

	c, err := loadConfig()
	if err != nil {
		return err
	}
	fabric.config = c

	log.Debugf("Fabric Configuration '%s' Loaded...", c.Metadata.Name)
	log.Debugf("  Management Ports: %d", c.ManagementPortCount)
	log.Debugf("  Upstream Ports:   %d", c.UpstreamPortCount)
	log.Debugf("  Downstream Ports: %d", c.DownstreamPortCount)
	for _, switchConf := range c.Switches {
		log.Debugf("  Switch %s Configuration: %s", switchConf.Id, switchConf.Metadata.Name)
		log.Debugf("    Management Ports: %d", switchConf.ManagementPortCount)
		log.Debugf("    Upstream Ports:   %d", switchConf.UpstreamPortCount)
		log.Debugf("    Downstream Ports: %d", switchConf.DownstreamPortCount)
	}

	f.switches = make([]Switch, len(c.Switches))
	for switchIdx, switchConf := range c.Switches {
		log.Infof("Initialize switch %s", switchConf.Id)
		f.switches[switchIdx] = Switch{
			id:     switchConf.Id,
			fabric: f,
			config: &c.Switches[switchIdx],
			ports:  make([]Port, len(switchConf.Ports)),
		}

		s := &f.switches[switchIdx]

		log.Infof("identify switch %s", switchConf.Id)
		if err := s.identify(); err != nil {
			log.WithError(err).Warnf("Failed to identify switch %s", s.id)
		}
		
		log.Infof("Switch %s identified: PAX %d", switchConf.Id, s.paxId)

		for portIdx, portConf := range switchConf.Ports {
			portType := portConf.getPortType()

			s.ports[portIdx] = Port{
				swtch:  &f.switches[switchIdx],
				config: &switchConf.Ports[portIdx],
				typ:    portType,
				idx:    portIdx,
			}
		}
	}

	/*
		// create the endpoints

		// Endpoint and Port relation
		//
		//       Endpoint         Port           Switch
		// [0  ] Rabbit           Mgmt           0, 1              One endpoint per mgmt (one mgmt port per switch)
		// [1  ] Compute 0        USP0			 0                 One endpoint per compuete
		// [2  ] Compute 1        USP1           0
		//   ...
		// [N-1] Compute N        USPN           1
		// [N  ] Drive 0 PF       DSP0           0 ---------------|
		// [N+1] Drive 0 VF0      DSP0           0                | Each drive is enumerated out to M endpoints
		// [N+2] Drive 0 VF1      DSP0           0                |   1 for the physical function (unused)
		//   ...                                                  |   1 for the rabbit
		// [N+M] Drive 0 VFM-1    DSP0           0 ---------------|   1 per compute
		//

			f.managementEndpointCount = 1
			f.upstreamEndpointCount = f.config.UpstreamPortCount

			mangementAndUpstreamEndpointCount := f.managementEndpointCount + f.upstreamEndpointCount
			f.downstreamEndpointCount = (1 // PF
				 + mangementAndUpstreamEndpointCount) * f.config.DownstreamPortCount

			f.endpoints = make([]Endpoint, mangementAndUpstreamEndpointCount+f.downstreamEndpointCount)

			for endpointIdx := range f.endpoints {
				endpoint := &f.endpoints[endpointIdx]

				endpoint.id = strconv.Itoa(endpointIdx)

				switch {
				case f.isManagementEndpoint(endpointIdx):
					endpoint.endpointType = sf.PROCESSOR_EV150ET
					endpoint.ports = make([]*Port, len(fabric.switches))
					for switchIdx, s := range fabric.switches {
						port := s.findPort(sf.MANAGEMENT_PORT_PV130PT, 0)

						endpoint.ports[switchIdx] = port

						port.endpoints = make([]*Endpoint, 1)
						port.endpoints[0] = endpoint
					}
				case f.isUpstreamEndpoint(endpointIdx):
					port := f.findPort(sf.UPSTREAM_PORT_PV130PT, f.getUpstreamEndpointRelativePortIndex(endpointIdx))

					endpoint.endpointType = sf.STORAGE_INITIATOR_EV150ET
					endpoint.ports = make([]*Port, 1)
					endpoint.ports[0] = port

					port.endpoints = make([]*Endpoint, 1)
					port.endpoints[0] = endpoint

				case f.isDownstreamEndpoint(endpointIdx):
					port := f.findPort(sf.DOWNSTREAM_PORT_PV130PT, f.getDownstreamEndpointRelativePortIndex(endpointIdx))

					endpoint.endpointType = sf.DRIVE_EV150ET
					endpoint.ports = make([]*Port, 1)
					endpoint.ports[0] = port

					if len(port.endpoints) == 0 {
						port.endpoints = make([]*Endpoint, 1 // PF
							 +mangementAndUpstreamEndpointCount)
						// we will initialize the port's endpoints when the endpointGroup is initialized
					}

				default:
					panic(fmt.Errorf("Unhandled endpoint index %d", endpointIdx))
				}
			}
	*/

	/*
		// create the endpoint groups & connections

		f.endpointGroups = make([]EndpointGroup, mangementAndUpstreamEndpointCount)
		f.connections = make([]Connection, mangementAndUpstreamEndpointCount)
		for endpointGroupIdx := range fabric.endpointGroups {
			endpointGroup := &fabric.endpointGroups[endpointGroupIdx]
			connection := &fabric.connections[endpointGroupIdx]

			endpointGroup.id = strconv.Itoa(endpointGroupIdx)
			endpointGroup.connection = connection
			endpointGroup.endpoints = make([]*Endpoint, 1+f.config.DownstreamPortCount)
			endpointGroup.endpoints[0] = &f.endpoints[endpointGroupIdx] // Mgmt or USP

			for idx := range endpointGroup.endpoints[1:] {
				endpointGroup.endpoints[1+idx] = &f.endpoints[endpointGroupIdx+mangementAndUpstreamEndpointCount+idx*(mangementAndUpstreamEndpointCount)]
			}

			connection.endpointGroup = endpointGroup

		}

		// initialize ports

		for _, s := range f.switches {
			for _, p := range s.ports {
				if err := p.Initialize(); err != nil {
					log.WithError(err).Errorf("Switch %s Port %s failed to initialize", s.id, p.id)
				}
			}
		}

		// initialize connections

		for _, c := range f.connections {
			if err := c.Initialize(); err != nil {
				log.WithError(err).Errorf("Connection %s failed to initialize", c.endpointGroup.id)
			}
		}
	*/

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

	s, err := fabric.findSwitch(switchId)
	if err != nil {
		return ec.ErrNotFound
	}

	model.Id = switchId
	model.SwitchType = sf.PC_IE_PP

	model.Status = s.getStatus()
	model.Model = s.getModel()
	model.Manufacturer = s.getManufacturer()
	model.SerialNumber = s.getSerialNumber()
	model.FirmwareVersion = s.getFirmwareVersion()

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
	model.PortType = p.typ
	model.PortId = strconv.Itoa(p.config.Port)

	model.Width = int64(p.config.Width)
	model.ActiveWidth = 0 // TODO

	//model.MaxSpeedGbps = 0 // TODO
	//model.CurrentSpeedGbps = 0 // TODO

	model.LinkState = sf.ENABLED_PV130LST
	//model.LinkStatus = // TODO

	model.Links.AssociatedEndpointsodataCount = int64(len(p.endpoints))
	model.Links.AssociatedEndpoints = make([]sf.OdataV4IdRef, model.Links.AssociatedEndpointsodataCount)
	for idx, ep := range p.endpoints {
		model.Links.AssociatedEndpoints[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, ep.id)
	}

	return nil
}

// FabricIdEndpointsGet -
func FabricIdEndpointsGet(fabricId string, model *sf.EndpointCollectionEndpointCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.endpoints))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx := range fabric.endpoints {
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

	type Oem struct {
		Pdfid         int
		Bound         bool
		BoundPaxId    int
		BoundHvdPhyId int
		BoundHvdLogId int
	}

	oem := Oem{
		Pdfid:         int(ep.pdfid),
		Bound:         ep.bound,
		BoundPaxId:    int(ep.boundPaxId),
		BoundHvdPhyId: int(ep.boundHvdPhyId),
		BoundHvdLogId: int(ep.boundHvdLogId),
	}

	model.Oem = openapi.MarshalOem(oem)

	return nil
}

func FabricIdEndpointGroupsGet(fabricId string, model *sf.EndpointGroupCollectionEndpointGroupCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.endpointGroups))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx := range fabric.endpointGroups {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/EndpointGroups/%d", fabricId, idx)
	}

	return nil
}

func FabricIdEndpointGroupsEndpointIdGet(fabricId string, groupId string, model *sf.EndpointGroupV130EndpointGroup) error {
	if !isFabric(fabricId) || !isEndpointGroup(groupId) {
		return ec.ErrNotFound
	}

	endpointGroup, err := fabric.findEndpointGroup(groupId)
	if err != nil {
		return err
	}

	model.Links.EndpointsodataCount = int64(len(endpointGroup.endpoints))
	model.Links.Endpoints = make([]sf.OdataV4IdRef, model.Links.EndpointsodataCount)
	for idx, ep := range endpointGroup.endpoints {
		model.Links.Endpoints[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, ep.id)
	}

	model.Links.ConnectionsodataCount = 1
	model.Links.Connections = make([]sf.OdataV4IdRef, model.Links.ConnectionsodataCount)
	model.Links.Connections[0].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections/%s", fabricId, endpointGroup.id)

	return nil
}

// FabricIdConnectionsGet -
func FabricIdConnectionsGet(fabricId string, model *sf.ConnectionCollectionConnectionCollection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fabric.connections))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, c := range fabric.connections {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections/%s", fabricId, c.endpointGroup.id)
	}

	return nil
}

// FabricIdConnectionsConnectionIdGet
func FabricIdConnectionsConnectionIdGet(fabricId string, connectionId string, model *sf.ConnectionV100Connection) error {
	if !isFabric(fabricId) || !isConnection(connectionId) {
		return ec.ErrNotFound
	}

	connection, err := fabric.findConnection(connectionId)
	if err != nil {
		return err
	}

	endpointGroup := connection.endpointGroup

	model.Id = connectionId
	model.ConnectionType = sf.STORAGE_CV100CT

	model.Links.InitiatorEndpointsodataCount = 1
	model.Links.InitiatorEndpoints = make([]sf.OdataV4IdRef, model.Links.InitiatorEndpointsodataCount)
	model.Links.InitiatorEndpoints[0].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, endpointGroup.endpoints[0].id)

	model.Links.TargetEndpointsodataCount = int64(len(endpointGroup.endpoints) - 1)
	model.Links.TargetEndpoints = make([]sf.OdataV4IdRef, model.Links.TargetEndpointsodataCount)
	for idx, endpoint := range endpointGroup.endpoints[1:] {
		model.Links.TargetEndpoints[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, endpoint.id)
	}

	// TODO: Fill out connection.VolumeInfo[] ConnectionV100VolumeInfo

	return nil
}
