package nvmenamespace

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	. "stash.us.cray.com/rabsw/nnf-ec/internal/common"
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/nvme"
)

const (
	ResourceBlockId = "Rabbit"
)

const (
	defaultStoragePoolId = "0"
)

// Manager -
type Manager struct {
	id   string
	ctrl NvmeControllerInterface

	config *ConfigFile

	storage []Storage
}

// Storage -
type Storage struct {
	id      string
	address string

	config *ControllerConfig

	virtManagementEnabled bool
	controllers           []StorageController
	volumes               []Volume

	// These values allow us to communicate a storage device with its corresponding
	// Fabric Controller
	fabricId string
	switchId string
	portId   string

	device NvmeDeviceInterface
}

// StorageController -
type StorageController struct {
	id             string
	controllerId   uint16
	functionNumber uint16

	deviceCtrl NvmeDeviceControllerInterface
}

// Volumes -
type Volume struct {
	id          string
	namespaceId nvme.NamespaceIdentifier
}

// TODO: We may want to put this manager under a resource block
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId} // <- Rabbit
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId} // <- Also Rabbit & Computes
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId}/​PCIeDevices/​{PCIeDeviceId}
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId}/​PCIeDevices/​{PCIeDeviceId}/​PCIeFunctions/{​PCIeFunctionId}
//
//   /​redfish/​v1/​ResourceBlocks/{​ResourceBlockId}/​Systems/{​ComputerSystemId}/​Storage/​{StorageId}/​Controllers/​{ControllerId}

var mgr Manager

func init() {
	RegisterNvmeInterface(&mgr)
}

func isStorageId(id string) bool     { _, err := mgr.findStorage(id); return err == nil }
func isStoragePoolId(id string) bool { return id == defaultStoragePoolId }

func (m *Manager) findStorage(storageId string) (*Storage, error) {
	id, err := strconv.Atoi(storageId)
	if err != nil {
		return nil, ec.ErrNotFound
	}

	if !(id < len(m.storage)) {
		return nil, ec.ErrNotFound
	}

	return &m.storage[id], nil
}

func (m *Manager) findStorageController(storageId string, controllerId string) (*StorageController, error) {
	s, err := m.findStorage(storageId)
	if err != nil {
		return nil, err
	}

	return s.findController(controllerId)
}

// GetVolumes -
func (m *Manager) GetVolumes(controllerId string) ([]string, error) {
	volumes := []string{}
	for _, s := range m.storage {
		c, err := s.findController(controllerId)
		if err != nil {
			return volumes, err
		}

		nsids, err := s.device.ListNamespaces(c.functionNumber)
		if err != nil {
			return volumes, err
		}

		for _, nsid := range nsids {
			for _, v := range s.volumes {
				if v.namespaceId == nsid {
					volumes = append(volumes, fmt.Sprintf("/redfish/v1/Storage/%s/Volumes/%s", s.id, v.id))
				}
			}
		}

	}

	return volumes, nil
}

// AttachVolumeToStorageController - Will attempt to decode the (StorageId, VolumeId) defined in
// the odataid string and attach that volume to the given controller.
func (m *Manager) AttachVolume(odataid string, controllerId string) error {

	// Wish I could use the router for this one, but all the hooks seem trapped to an http.Request
	// so manual extraction is needed. TODO: Add this to a unit test.
	fields := strings.Split(odataid, "/")
	if len(fields) != len(strings.Split("/redfish/v1/Storage/0/Volumes/0", "/")) {
		return ec.ErrBadRequest
	}

	storageId := fields[4]
	volumeId := fields[6]

	s, err := m.findStorage(storageId)
	if err != nil {
		return ec.ErrBadRequest
	}

	v, err := s.findVolume(volumeId)
	if err != nil {
		return ec.ErrBadRequest
	}

	c, err := s.findController(controllerId)
	if err != nil {
		return ec.ErrBadRequest
	}

	// TODO: Check if volume is already attached to the storage controller and return failure

	if err := s.device.AttachNamespace(v.namespaceId, c.controllerId); err != nil {
		log.WithError(err).Errorf("Attach namespace failed")
		return ec.ErrInternalServerError
	}

	return nil
}

func (s *Storage) initialize(conf *ControllerConfig, device string) error {
	s.address = device
	s.config = conf

	return nil
}

func (s *Storage) findController(controllerId string) (*StorageController, error) {
	for idx, ctrl := range s.controllers {
		if ctrl.id == controllerId {
			return &s.controllers[idx], nil
		}
	}

	return nil, ec.ErrNotFound
}

func (s *Storage) getStatus() (stat sf.ResourceStatus) {
	if len(s.controllers) == 0 {
		stat.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		stat.Health = sf.OK_RH
		stat.State = sf.ENABLED_RST
	}

	return stat
}

func (s *Storage) createVolume(sizeInBytes uint64) (string, error) {
	namespaceId, err := s.device.CreateNamespace(sizeInBytes)
	if err != nil {
		return "", err
	}

	id := strconv.Itoa(int(namespaceId))
	s.volumes = append(s.volumes, Volume{
		id:          id,
		namespaceId: namespaceId,
	})

	sort.Slice(s.volumes, func(i, j int) bool {
		return s.volumes[i].id < s.volumes[j].id
	})

	return id, nil
}

func (s *Storage) deleteVolume(volumeId string) error {
	for idx, volume := range s.volumes {
		if volume.id == volumeId {
			if err := s.device.DeleteNamespace(volume.namespaceId); err != nil {
				return ec.ErrInternalServerError
			}

			// remove the volume from the array
			copy(s.volumes[idx:], s.volumes[idx+1:]) // shift left 1 at idx
			s.volumes = s.volumes[:len(s.volumes)-1] // truncate tail

			return nil
		}
	}

	return ec.ErrNotFound
}

func (s *Storage) findVolume(volumeId string) (*Volume, error) {
	for idx, v := range s.volumes {
		if v.id == volumeId {
			return &s.volumes[idx], nil
		}
	}

	return nil, ec.ErrNotFound
}

// Initialize
func Initialize(ctrl NvmeControllerInterface) error {

	mgr = Manager{
		id:   ResourceBlockId,
		ctrl: ctrl,
	}

	log.SetLevel(log.DebugLevel)

	log.Infof("Initialize %s NVMe Namespace Manager", mgr.id)

	conf, err := loadConfig()
	if err != nil {
		log.WithError(err).Errorf("Failed to load %s configuration", mgr.id)
		return err
	}

	mgr.config = conf

	log.Debugf("NVMe Configuration '%s' Loaded...", conf.Metadata.Name)
	log.Debugf("  Controller Config:")
	log.Debugf("    Virtual Functions: %d", conf.Storage.Controller.Functions)
	log.Debugf("    Num Resources: %d", conf.Storage.Controller.Resources)
	log.Debugf("  Device List: %+v", conf.Storage.Devices)

	mgr.storage = make([]Storage, len(conf.Storage.Devices))
	for storageIdx, storageConfig := range conf.Storage.Devices {
		storage := &mgr.storage[storageIdx]

		storage.id = strconv.Itoa(storageIdx)
		if err := storage.initialize(&conf.Storage.Controller, storageConfig); err != nil {
			log.WithError(err).Errorf("Failed to initialize storage device %s", storage.id)
		}
	}

	PortEventManager.Subscribe(PortEventSubscriber{
		HandlerFunc: PortEventHandler,
		Data:        &mgr,
	})

	return nil
}

// PortEventHandler - Receives port events from the event manager
func PortEventHandler(event PortEvent, data interface{}) {
	m := data.(*Manager)

	log.Infof("%s Port Event Received %+v", m.id, event)

	if event.PortType != PORT_TYPE_DSP {
		return
	}

	// Storage ID is related to the fabric controller that is running. Convert the
	// PortEvent to our storage index.
	idx, err := FabricController.ConvertPortEventToRelativePortIndex(event)
	if err != nil {
		log.WithError(err).Errorf("Unable to find port index for event %+v", event)
		return
	}

	if !(idx < len(m.storage)) {
		log.Errorf("No storage device exists for index %d", idx)
		return
	}

	s := &m.storage[idx]

	switch event.EventType {
	case PORT_EVENT_UP:

		// Connect
		device, err := m.ctrl.NewNvmeDevice(event.FabricId, event.SwitchId, event.PortId)
		if err != nil {
			log.WithError(err).Errorf("Could not allocate storage controller")
			return
		}

		s.device = device
		s.virtManagementEnabled, err = device.IsVirtualizationManagement()

		// Initialize a Storage object with the required number of controllers
		initFunc := func(s *Storage) SecondaryControllersInitFunc {
			return func(count uint8) {
				if count > uint8(s.config.Functions) {
					count = uint8(s.config.Functions)
				}

				s.controllers = make([]StorageController, 1 /*PF*/ +count)
			}
		}

		handlerFunc := func(s *Storage) SecondaryControllerHandlerFunc {
			return func(controllerId uint16, controllerOnline bool, virtualFunctionNumber uint16, numVQResourcesAssinged, numVIResourcesAssigned uint16) error {
				if !(controllerId < uint16(len(s.controllers))) {
					return nil
				}

				s.controllers[int(controllerId)] = StorageController{
					id:             strconv.Itoa(int(controllerId)),
					controllerId:   controllerId,
					functionNumber: virtualFunctionNumber,
				}

				if !s.virtManagementEnabled {
					return nil
				}

				if numVQResourcesAssinged != uint16(s.config.Resources) {
					if err := s.device.AssignControllerResources(controllerId, VQResourceType, uint32(int(numVQResourcesAssinged)-s.config.Resources)); err != nil {
						return err
					}
				}

				if numVIResourcesAssigned != uint16(s.config.Resources) {
					if err := s.device.AssignControllerResources(controllerId, VIResourceType, uint32(int(numVIResourcesAssigned)-s.config.Resources)); err != nil {
						return err
					}
				}

				if !controllerOnline {
					if err := s.device.OnlineController(controllerId); err != nil {
						return err
					}
				}

				return nil
			}
		}

		if err := device.EnumerateSecondaryControllers(initFunc(s), handlerFunc(s)); err != nil {
			log.WithError(err).Errorf("Failed to enumerate %s storage controllers ", s.id)
		}

		// Port is ready to make connections
		event.EventType = PORT_EVENT_READY
		PortEventManager.Publish(event)

	case PORT_EVENT_DOWN:
		// TODO: Set this and all controllers down
	}
}

// Get -
func Get(model *sf.StorageCollectionStorageCollection) error {
	model.MembersodataCount = int64(len(mgr.storage))
	model.Members = make([]sf.OdataV4IdRef, int(model.MembersodataCount))
	for idx, s := range mgr.storage {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s", s.id)
	}
	return nil
}

// StorageIdGet -
func StorageIdGet(storageId string, model *sf.StorageV190Storage) error {

	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	model.Status = s.getStatus()

	// TODO: The model is missing a bunch of stuff
	// Manufacturer, Model, PartNumber, SerialNumber, etc.

	model.Controllers.OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/Controllers", storageId)
	model.StoragePools.OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StoragePools", storageId)
	model.Volumes.OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/Volumes", storageId)

	return nil
}

// StorageIdStoragePoolsGet -
func StorageIdStoragePoolsGet(storageId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StoragePool/%s", storageId, defaultStoragePoolId)

	return nil
}

// StorageIdStoragePoolIdGet -
func StorageIdStoragePoolIdGet(storageId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	if !isStorageId(storageId) || !isStoragePoolId(storagePoolId) {
		return ec.ErrNotFound
	}

	// TODO: This should reflect the total namespaces allocated over the drive
	model.Capacity = sf.CapacityV100Capacity{
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes:   0,
			ConsumedBytes:    0,
			GuaranteedBytes:  0,
			ProvisionedBytes: 0,
		},
	}

	// TODO
	model.RemainingCapacityPercent = 0

	return nil
}

// StorageIdControllersGet -
func StorageIdControllersGet(storageId string, model *sf.StorageControllerCollectionStorageControllerCollection) error {

	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	model.MembersodataCount = int64(len(s.controllers))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, c := range s.controllers {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/Controllers/%s", storageId, c.id)
	}

	return nil
}

// StorageIdControllerIdGet -
func StorageIdControllerIdGet(storageId, controllerId string, model *sf.StorageControllerV100StorageController) error {

	_, err := mgr.findStorageController(storageId, controllerId)
	if err != nil {
		return err
	}

	// Fill in the relative endpoint for this storage controller
	endpointId, err := FabricController.FindDownstreamEndpoint(storageId, controllerId)
	if err != nil {
		return err
	}

	model.Links.EndpointsodataCount = 1
	model.Links.Endpoints = make([]sf.OdataV4IdRef, model.Links.EndpointsodataCount)
	model.Links.Endpoints[0].OdataId = endpointId

	// model.Links.PCIeFunctions

	/*
		f := sf.PcIeFunctionV123PcIeFunction{
			ClassCode: "",
			DeviceClass: "",
			DeviceId: "",
			VendorId: "",
			SubsystemId: "",
			SubsystemVendorId: "",
			FunctionId: 0,
			FunctionType: sf.PHYSICAL_PCIFV123FT, // or sf.VIRTUAL_PCIFV123FT
			Links: sf.PcIeFunctionV123Links {
				StorageControllersodataCount: 1,
				StorageControllers: make([]sf.StorageStorageController, 1),
			},
		}
	*/

	model.NVMeControllerProperties = sf.StorageControllerV100NvMeControllerProperties{
		ControllerType: sf.IO_SCV100NVMCT, // OR ADMIN IF PF
	}

	return nil
}

// StorageIdVolumesGet -
func StorageIdVolumesGet(storageId string, model *sf.VolumeCollectionVolumeCollection) error {
	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	// TODO: If s.ctrl is down - fail

	model.MembersodataCount = int64(len(s.volumes))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, volume := range s.volumes {
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/Volumes/%s", storageId, volume.id)
	}

	return nil
}

// StorageIdVolumeIdGet -
func StorageIdVolumeIdGet(storageId, volumeId string, model *sf.VolumeV161Volume) error {
	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	v, err := s.findVolume(volumeId)
	if err != nil {
		return err
	}

	// TODO: If s.ctrl is down - fail

	ns, err := s.device.GetNamespace(nvme.NamespaceIdentifier(v.namespaceId))
	if err != nil {
		return ec.ErrNotFound
	}

	formatGUID := func(guid []byte) string {
		var b strings.Builder
		for _, byt := range guid {
			b.WriteString(fmt.Sprintf("%02x", byt))
		}
		return b.String()
	}

	lbaFormat := ns.LBAFormats[ns.FormattedLBASize.Format]
	blockSizeInBytes := int64(math.Pow(2, float64(lbaFormat.LBADataSize)))

	model.BlockSizeBytes = blockSizeInBytes
	model.CapacityBytes = int64(ns.Capacity) * blockSizeInBytes
	model.Id = v.id
	model.Identifiers = make([]sf.ResourceIdentifier, 2)
	model.Identifiers = []sf.ResourceIdentifier{
		{
			DurableNameFormat: sf.NSID_RV1100DNF,
			DurableName:       fmt.Sprintf("%d", v.namespaceId),
		},
		{
			DurableNameFormat: sf.NGUID_RV1100DNF,
			DurableName:       formatGUID(ns.GloballyUniqueIdentifier[:]),
		},
	}

	model.Capacity = sf.CapacityV100Capacity{
		IsThinProvisioned: ns.Features.Thinp == 1,
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes: int64(ns.Capacity) * blockSizeInBytes,
			ConsumedBytes:  int64(ns.Utilization) * blockSizeInBytes,
		},
	}

	model.NVMeNamespaceProperties = sf.VolumeV161NvMeNamespaceProperties{
		FormattedLBASize:                  fmt.Sprintf("%d", model.BlockSizeBytes),
		IsShareable:                       ns.MultiPathIOSharingCapabilities.Sharing == 1,
		MetadataTransferredAtEndOfDataLBA: lbaFormat.MetadataSize != 0,
		NamespaceId:                       fmt.Sprintf("%d", v.namespaceId),
		NumberLBAFormats:                  int64(ns.NumberOfLBAFormats),
	}

	model.VolumeType = sf.RAW_DEVICE_VVT

	// TODO: Find the attached status of the volume - if it is attached via a connection
	// to an endpoint that should go in model.Links.ClientEndpoints or model.Links.ServerEndpoints

	// TODO: Maybe StorageGroups??? An array of references to Storage Groups that includes this volume.
	// Storage Groups could be the Rabbit Slice

	// TODO: Should reference the Storage Pool

	return nil
}

// StorageIdVolumePost -
func StorageIdVolumePost(storageId string, model *sf.VolumeV161Volume) error {
	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	volumeId, err := s.createVolume(uint64(model.CapacityBytes))

	// TODO: We should parse the error and make it more obvious (404, 405, etc)
	if err != nil {
		return err
	}

	return StorageIdVolumeIdGet(storageId, volumeId, model)
}

// StorageIdVolumeIdDelete -
func StorageIdVolumeIdDelete(storageId, volumeId string) error {
	s, err := mgr.findStorage(storageId)
	if err != nil {
		return err
	}

	return s.deleteVolume(volumeId)
}
