package fabric

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	. "stash.us.cray.com/rabsw/nnf-ec/internal/common"
	openapi "stash.us.cray.com/sp/rfsf-openapi/pkg/common"
	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
	"stash.us.cray.com/~roiger/switchtec-fabric/pkg/switchtec"
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
	id  string
	idx int

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
	id       string
	fabricId string // This is the absolute ID within the fabricP

	portType   sf.PortV130PortType
	linkStatus sf.PortV130LinkStatus // TODO: Link Width, Link Speed etc

	swtch  *Switch
	config *PortConfig

	endpoints []*Endpoint
}

type Endpoint struct {
	id  string
	idx int

	endpointType sf.EndpointV150EntityType
	controllerId string // For Target Endpoints, this represents the VF

	ports []*Port

	// OEM fields -  marshalled?
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

	initiator **Endpoint
}

type Connection struct {
	endpointGroup *EndpointGroup
	volumes       []VolumeInfo
}

type VolumeInfo struct {
	odataid string
}

var fabric Fabric

func init() {
	RegisterFabricController(&fabric)
}

func isFabric(id string) bool        { return id == FabricId }
func isSwitch(id string) bool        { _, err := fabric.findSwitch(id); return err == nil }
func isEndpoint(id string) bool      { _, err := fabric.findEndpoint(id); return err == nil }
func isEndpointGroup(id string) bool { _, err := fabric.findEndpointGroup(id); return err == nil }
func isConnection(id string) bool    { _, err := fabric.findConnection(id); return err == nil }

// GetSwitchtecDevice -
func (f *Fabric) GetSwitchtecDevice(fabricId, switchId string) (*switchtec.Device, error) {
	if fabricId != f.id {
		return nil, fmt.Errorf("Fabric %s not found", fabricId)
	}

	s, err := f.findSwitch(switchId)
	if err != nil {
		return nil, err
	}

	return s.dev.Get(), nil
}

// GetPortPDFID
func (f *Fabric) GetPortPDFID(fabricId, switchId, portId string, controllerId uint16) (uint16, error) {
	if fabricId != f.id {
		return 0, fmt.Errorf("Fabric %s not found", fabricId)
	}

	p, err := f.findSwitchPort(switchId, portId)
	if err != nil {
		return 0, err
	}

	if p.portType != sf.DOWNSTREAM_PORT_PV130PT {
		return 0, fmt.Errorf("Port %s of Type %s has no PDFID", portId, p.portType)
	}

	if !(int(controllerId) < len(p.endpoints)) {
		return 0, fmt.Errorf("Controller ID beyond available port endpoints")
	}

	return p.endpoints[int(controllerId)].pdfid, nil
}

// ConvertPortEventToRelativePortIndex
func (f *Fabric) ConvertPortEventToRelativePortIndex(event PortEvent) (int, error) {
	if event.FabricId != f.id {
		return -1, fmt.Errorf("Fabric %s not found for event %+v", event.FabricId, event)
	}

	var idx = 0
	for _, s := range f.switches {
		for _, p := range s.ports {

			var correctType = false
			switch event.PortType {
			case USP_PORT_TYPE:
				correctType = (p.portType == sf.MANAGEMENT_PORT_PV130PT ||
					p.portType == sf.UPSTREAM_PORT_PV130PT)
			case DSP_PORT_TYPE:
				correctType = p.portType == sf.DOWNSTREAM_PORT_PV130PT

			}

			if correctType == true {
				if s.id == event.SwitchId && p.id == event.PortId {
					return idx, nil
				}

				idx++
			}
		}
	}

	return -1, fmt.Errorf("Relative Port Index not found for event %+v", event)
}

// FindDownstreamEndpoint -
func (f *Fabric) FindDownstreamEndpoint(portId, functionId string) (string, error) {
	idx, err := strconv.Atoi(portId)
	if err != nil {
		return "", ec.ErrNotFound
	}
	port := f.findPortByType(sf.DOWNSTREAM_PORT_PV130PT, idx)
	if port == nil {
		return "", ec.ErrNotFound
	}
	ep := port.findEndpoint(functionId)
	if ep == nil {
		return "", ec.ErrNotFound
	}

	return fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", f.id, ep.id), nil
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

// findPortByType - Finds the i'th port of portType in the fabric
func (f *Fabric) findPortByType(portType sf.PortV130PortType, idx int) *Port {
	switch portType {
	case sf.MANAGEMENT_PORT_PV130PT:
		return f.switches[idx].findPortByType(portType, 0)
	case sf.UPSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.config.UpstreamPortCount {
				return s.findPortByType(portType, idx)
			}
			idx = idx - s.config.UpstreamPortCount
		}
	case sf.DOWNSTREAM_PORT_PV130PT:
		for _, s := range f.switches {
			if idx < s.config.DownstreamPortCount {
				return s.findPortByType(portType, idx)
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
	}

	return fmt.Errorf("Identify Switch %s: Could Not ID Switch", s.id) // TODO: Switch not found
}

func (s *Switch) getStatus() (stat sf.ResourceStatus) {

	if s.dev == nil {
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

func (s *Switch) findPort(portId string) (*Port, error) {
	for _, p := range s.ports {
		if p.id == portId {
			return &p, nil
		}
	}

	return nil, ec.ErrNotFound
}

// findPort - Finds the i'th port of portType in the switch
func (s *Switch) findPortByType(portType sf.PortV130PortType, idx int) *Port {
	for portIdx, port := range s.ports {
		if port.portType == portType {
			if idx == 0 {
				return &s.ports[portIdx]
			}
			idx--
		}
	}

	panic(fmt.Sprintf("Switch Port %d Not Found", idx))
}

func (s *Switch) isDown() bool {
	return !s.isReady()
}

func (p *Port) GetBaseEndpointIdx() int { return p.endpoints[0].idx }

func (p *Port) findEndpoint(functionId string) *Endpoint {
	id, err := strconv.Atoi(functionId)
	if err != nil {
		return nil
	}
	if !(id < len(p.endpoints)) {
		return nil
	}
	return p.endpoints[id]
}

func (p *Port) LinkStatus() error {
	// TODO
	p.linkStatus = sf.LINK_UP_PV130LS
	return nil
}

func (p *Port) Initialize() error {
	log.Infof("Initialize port %s: %s %d", p.id, p.config.Name, p.config.Port)

	if p.swtch.isDown() {
		log.Warnf("port %s switch is down", p.id)
		return nil
	}

	if err := p.LinkStatus(); err != nil {
		return err
	}

	switch p.portType {
	case sf.DOWNSTREAM_PORT_PV130PT:

		processPort := func(port *Port) func(*switchtec.DumpEpPortDevice) error {
			return func(epPort *switchtec.DumpEpPortDevice) error {

				if switchtec.EpPortType(epPort.Hdr.Typ) != switchtec.DeviceEpPortType {
					log.Errorf("Port %s is down", p.id)
					// Port & Associated Endpoints are Down/Unreachable
					//p.Down() // TODO
				}

				//log.Debugf("Processing EP Functions: %d", epPort.Ep.Functions)
				for idx, f := range epPort.Ep.Functions {

					if idx >= len(p.endpoints) {
						break
					}

					ep := p.endpoints[idx]
					ep.controllerId = strconv.Itoa(int(f.VFNum))

					ep.pdfid = f.PDFID
					ep.bound = f.Bound != 0
					ep.boundPaxId = f.BoundPAXID
					ep.boundHvdPhyId = f.BoundHVDPhyPID
					ep.boundHvdLogId = f.BoundHVDLogPID
				}

				return nil
			}
		}

		log.Infof("Switch %s enumerting DSP %d", p.swtch.id, p.config.Port)
		if err := p.swtch.dev.EnumerateEndpoint(uint8(p.config.Port), processPort(p)); err != nil {
			return err
		}
	}

	return nil
}

// Initialize - Will create the bindings between initiator and downstream ports
// An Endpoint Group exists for every virtual domain, representing the logical domain
// for initators and there downstream ports. We perform a BIND operation on the switch
// to form a connection between initiator and DSP.
func (c *Connection) Initialize() error {

	endpointGroup := c.endpointGroup
	initiatorEndpoint := *(c.endpointGroup.initiator)

	if initiatorEndpoint != c.endpointGroup.endpoints[0] {
		panic("Initiator endpoint must be at index 0 of an endpoint group to support connection algorithm")
	}

	if initiatorEndpoint.endpointType == sf.PROCESSOR_EV150ET {
		if len(initiatorEndpoint.ports) != 2 {
			panic("Processor endpoint expected to have two ports for implemented connection algorithm")
		}
	}

	// Iterative over all DSP
	for epIdx, ep := range endpointGroup.endpoints[1:] {

		if ep.endpointType != sf.DRIVE_EV150ET {
			panic("Expected drive endpoint type for connection")
		}

		if len(ep.ports) != 1 {
			panic("Expected port count to be one for Drive Endpoint")
		}

		for _, port := range initiatorEndpoint.ports {

			if ep.ports[0] == port {
				swtch := port.swtch
				dev := swtch.dev

				logicalPortId := uint8(epIdx)
				if initiatorEndpoint.endpointType == sf.PROCESSOR_EV150ET {
					logicalPortId -= uint8(swtch.idx * swtch.config.DownstreamPortCount)
				}

				if err := dev.Bind(uint8(port.config.Port), uint8(logicalPortId), ep.pdfid); err != nil {
					log.WithError(err).Warnf("Failed to bind port")
				}
			}
		}
	}

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
		log.WithError(err).Errorf("Failed to load % configuration", f.id)
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
	var fabricPortId = 0
	for switchIdx, switchConf := range c.Switches {
		log.Infof("Initialize switch %s", switchConf.Id)
		f.switches[switchIdx] = Switch{
			id:     switchConf.Id,
			idx:    switchIdx,
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
				id:         strconv.Itoa(portIdx),
				fabricId:   strconv.Itoa(fabricPortId),
				swtch:      &f.switches[switchIdx],
				config:     &switchConf.Ports[portIdx],
				portType:   portType,
				linkStatus: sf.NO_LINK_PV130LS,
			}

			fabricPortId++
		}
	}

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
	f.downstreamEndpointCount = (1 + // PF
		mangementAndUpstreamEndpointCount) * f.config.DownstreamPortCount

	log.Debugf("Creating Endpoints:")
	log.Debugf("   Management Endpoints: % 3d", f.managementEndpointCount)
	log.Debugf("   Upstream Endpoints:   % 3d", f.upstreamEndpointCount)
	log.Debugf("   Downstream Endpoints: % 3d", f.downstreamEndpointCount)

	f.endpoints = make([]Endpoint, mangementAndUpstreamEndpointCount+f.downstreamEndpointCount)

	for endpointIdx := range f.endpoints {
		endpoint := &f.endpoints[endpointIdx]

		endpoint.id = strconv.Itoa(endpointIdx)
		endpoint.idx = endpointIdx

		switch {
		case f.isManagementEndpoint(endpointIdx):
			endpoint.endpointType = sf.PROCESSOR_EV150ET
			endpoint.ports = make([]*Port, len(fabric.switches))
			for switchIdx, s := range fabric.switches {
				port := s.findPortByType(sf.MANAGEMENT_PORT_PV130PT, 0)

				endpoint.ports[switchIdx] = port

				port.endpoints = make([]*Endpoint, 1)
				port.endpoints[0] = endpoint
			}
		case f.isUpstreamEndpoint(endpointIdx):
			port := f.findPortByType(sf.UPSTREAM_PORT_PV130PT, f.getUpstreamEndpointRelativePortIndex(endpointIdx))

			endpoint.endpointType = sf.STORAGE_INITIATOR_EV150ET
			endpoint.ports = make([]*Port, 1)
			endpoint.ports[0] = port

			port.endpoints = make([]*Endpoint, 1)
			port.endpoints[0] = endpoint

		case f.isDownstreamEndpoint(endpointIdx):
			port := f.findPortByType(sf.DOWNSTREAM_PORT_PV130PT, f.getDownstreamEndpointRelativePortIndex(endpointIdx))

			//log.Debugf("Processing DSP Endpoint %d: Port %s", endpointIdx, port.id)
			//log.Debugf("  Relative Port Index:           % 3d", f.getDownstreamEndpointRelativePortIndex(endpointIdx))

			endpoint.endpointType = sf.DRIVE_EV150ET
			endpoint.ports = make([]*Port, 1)
			endpoint.ports[0] = port

			if len(port.endpoints) == 0 {
				port.endpoints = make([]*Endpoint, 1+ // PF
					mangementAndUpstreamEndpointCount)
				port.endpoints[0] = endpoint
			} else {
				port.endpoints[endpointIdx-port.GetBaseEndpointIdx()] = endpoint
			}

		default:
			panic(fmt.Errorf("Unhandled endpoint index %d", endpointIdx))
		}
	}

	// create the endpoint groups & connections

	// An Endpoint Groups is created for each managment and upstream endpoints, with
	// the associated target endpoints linked to form the group. This is conceptually
	// equivalent to the Host Virtualization Domains that exist in the PAX Switch.

	// A Connection is made for every endpoint (also representing the HVD). Connections
	// contain the attached volumes. The two are linked.
	f.endpointGroups = make([]EndpointGroup, mangementAndUpstreamEndpointCount)
	f.connections = make([]Connection, mangementAndUpstreamEndpointCount)
	for endpointGroupIdx := range fabric.endpointGroups {
		endpointGroup := &fabric.endpointGroups[endpointGroupIdx]
		connection := &fabric.connections[endpointGroupIdx]

		endpointGroup.id = strconv.Itoa(endpointGroupIdx)

		endpointGroup.endpoints = make([]*Endpoint, 1+f.config.DownstreamPortCount)
		endpointGroup.initiator = &endpointGroup.endpoints[0]
		endpointGroup.endpoints[0] = &f.endpoints[endpointGroupIdx] // Mgmt or USP

		for idx := range endpointGroup.endpoints[1:] {
			endpointGroup.endpoints[1+idx] = &f.endpoints[endpointGroupIdx+mangementAndUpstreamEndpointCount+idx*(mangementAndUpstreamEndpointCount)]
		}

		endpointGroup.connection = connection
		connection.endpointGroup = endpointGroup
	}

	PortEventManager.Subscribe(PortEventSubscriber{
		HandlerFunc: PortEventHandler,
		Data:        f,
	})

	log.Infof("%s Fabric Initialization Finished", f.id)
	return nil
}

// Start -
func Start() error {
	f := fabric
	log.Infof("%s Fabric Starting", f.id)

	// Enumerate over the switch ports and report events to the event
	// manager
	for _, s := range f.switches {
		for _, p := range s.ports {

			if err := p.Initialize(); err != nil {
				log.WithError(err).Errorf("Switch %s Port %s failed to initialize", s.id, p.id)
			}

			event := PortEvent{
				FabricId: f.id,
				SwitchId: s.id,
				PortId:   p.id,
			}

			switch p.portType {
			case sf.UPSTREAM_PORT_PV130PT, sf.MANAGEMENT_PORT_PV130PT:
				event.PortType = USP_PORT_TYPE
			case sf.DOWNSTREAM_PORT_PV130PT:
				event.PortType = DSP_PORT_TYPE
			default: // Ignore unintersting port types
				continue
			}

			switch p.linkStatus {
			case sf.LINK_UP_PV130LS:
				event.EventType = PORT_EVENT_UP
			default:
				event.EventType = PORT_EVENT_DOWN
			}

			log.Infof("Publishing port event %+v", event)
			PortEventManager.Publish(event)
		}
	}

	// TODO: Start monitoring LinkUp/Down status within the switch

	return nil
}

// PortEventHandler -
func PortEventHandler(event PortEvent, data interface{}) {
	f := data.(*Fabric)

	port, err := f.findSwitchPort(event.SwitchId, event.PortId)
	if err != nil {
		log.WithError(err).Errorf("Could not locate switch port for event %+v", event)
		return
	}

	// TODO: Respond to Port Ready event that makes the connection
	// between a port's endpoints and its USP

	switch event.EventType {
	case PORT_EVENT_READY:
		if port.portType == sf.DOWNSTREAM_PORT_PV130PT {
			/*
				// TODO: Bind this port to its initiators
				for endpointIdx, endpoint := range port.endpoints {
					if endpoint.controllerId != "" {

					}
				}
			*/
		}

	case PORT_EVENT_DOWN:
		// Set the port down & and all its connections
	}

	// TODO: Cannot make bindings when under virtualization management

	for _, c := range f.connections {
		if err := c.Initialize(); err != nil {
			log.WithError(err).Errorf("Connection %s failed to initialize", c.endpointGroup.id)
		}
	}
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
	model.PortType = p.portType
	model.PortId = strconv.Itoa(p.config.Port)

	model.Width = int64(p.config.Width)
	model.ActiveWidth = 0 // TODO

	//model.MaxSpeedGbps = 0 // TODO
	//model.CurrentSpeedGbps = 0 // TODO

	model.LinkState = sf.ENABLED_PV130LST
	model.LinkStatus = p.linkStatus

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

	model.Id = ep.id
	model.EndpointProtocol = sf.PC_IE_PP
	model.ConnectedEntities = make([]sf.EndpointV150ConnectedEntity, 1)
	model.ConnectedEntities = []sf.EndpointV150ConnectedEntity{{
		EntityType: ep.endpointType,
		EntityRole: sf.TARGET_EV150ER, // TODO
	}}

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
	initiator := *endpointGroup.initiator


	model.Id = connectionId
	model.ConnectionType = sf.STORAGE_CV100CT

	model.Links.InitiatorEndpointsodataCount = 1
	model.Links.InitiatorEndpoints = make([]sf.OdataV4IdRef, model.Links.InitiatorEndpointsodataCount)
	model.Links.InitiatorEndpoints[0].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, initiator.id)

	model.Links.TargetEndpointsodataCount = int64(len(endpointGroup.endpoints) - 1)
	model.Links.TargetEndpoints = make([]sf.OdataV4IdRef, model.Links.TargetEndpointsodataCount)
	for idx, endpoint := range endpointGroup.endpoints[1:] {
		model.Links.TargetEndpoints[idx].OdataId = fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, endpoint.id)
	}

	volumes, err := NvmeInterface.GetVolumes(initiator.controllerId)
	if err != nil {
		return err
	}

	model.VolumeInfo = make([]sf.ConnectionV100VolumeInfo, len(volumes))
	for idx, volume := range volumes {
		v := &model.VolumeInfo[idx]

		v.Volume.OdataId = volume
		v.AccessState = sf.OPTIMIZED_CV100AST
		v.AccessCapabilities = []sf.ConnectionV100AccessCapability{
			sf.READ_CV100AC,
			sf.WRITE_CV100AC,
		}
	}

	return nil
}

// FabricIdConnectionsConnectionIdPatch -
func FabricIdConnectionsConnectionIdPatch(fabricId string, connectionId string, model *sf.ConnectionV100Connection) error {
	if !isFabric(fabricId) {
		return ec.ErrNotFound
	}

	c, err := fabric.findConnection(connectionId)
	if err != nil {
		return err
	}

	initiator := *c.endpointGroup.initiator

	for _, volumeInfo := range model.VolumeInfo {
		odataid := volumeInfo.Volume.OdataId
		if err := NvmeInterface.AttachVolume(odataid, initiator.controllerId); err != nil {
			return err
		}
	}

	return nil
}
