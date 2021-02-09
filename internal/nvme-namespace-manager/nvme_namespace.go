package nvmenamespace

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"

	. "stash.us.cray.com/rabsw/nnf-ec/internal/common"
	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

const (
	ResourceBlockId = "Rabbit"
)

const (
	defaultStoragePoolId = "1"
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

	path        string
	controllers []StorageController

	ctrl NvmeStorageControllerInterface
}

// StorageController -
type StorageController struct {
	id string
}

// TODO: We may want to put this manager under a resource block
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId} // <- Rabbit
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId} // <- Also Rabbit & Computes
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId}/​PCIeDevices/​{PCIeDeviceId}
//   /​redfish/​v1/​ResourceBlocks/​{ResourceBlockId}/​Systems/{​ComputerSystemId}/​PCIeDevices/​{PCIeDeviceId}/​PCIeFunctions/{​PCIeFunctionId}
//
//   /​redfish/​v1/​ResourceBlocks/{​ResourceBlockId}/​Systems/{​ComputerSystemId}/​Storage/​{StorageId}/​Controllers/​{ControllerId}

var mgr Manager

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

func findStorageController(storageId string, controllerId string) (*StorageController, error) {
	return nil, ec.ErrNotFound
}

func (s *Storage) initialize(conf ControllerConfig, device string) error {
	s.address = device

	// TODO: This should query the switchtec device for the nvme's PF/VF
	//       list (i.e. Primary/Secondary Controllers)
	s.controllers = make([]StorageController, conf.VirtualFunctions)
	for idx := range s.controllers {
		s.controllers[idx].id = strconv.Itoa(idx)
	}
	return nil
}

func (s *Storage) getStatus() (stat sf.ResourceStatus) {
	if s.path == "" {
		stat.State = sf.UNAVAILABLE_OFFLINE_RST
	} else {
		stat.Health = sf.OK_RH
		stat.State = sf.ENABLED_RST
	}

	return stat
}

func Initialize(ctrl NvmeControllerInterface) error {

	mgr = Manager{
		id:   ResourceBlockId,
		ctrl: ctrl,
	}

	log.SetLevel(log.DebugLevel)
	log.Infof("Initialize %s Resource Block Manager", mgr.id)

	conf, err := loadConfig()
	if err != nil {
		log.WithError(err).Errorf("Failed to load % configuration", mgr.id)
	}

	mgr.config = conf

	log.Debugf("NVMe Configuration '%s' Loaded...", conf.Metadata.Name)
	log.Debugf("  Controller Config:")
	log.Debugf("    VirtualFunctions: %d", conf.ControllerConfig.VirtualFunctions)
	log.Debugf("  Device List: %+v", conf.Devices)

	mgr.storage = make([]Storage, len(conf.Devices))
	for storageIdx, storageConfig := range conf.Devices {
		storage := &mgr.storage[storageIdx]

		storage.id = strconv.Itoa(storageIdx)
		if err := storage.initialize(conf.ControllerConfig, storageConfig); err != nil {
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

	if event.PortType != DSP_PORT_TYPE {
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
		s.ctrl, err = m.ctrl.NewNvmeStorageController(event.FabricId, event.SwitchId, event.PortId)

		// Issue an Identify to the PF - check if virtualization management is enabled

		// Issue a ListSecondary to the PF
		// 		foreach (secondaryController):
		//			Add to Storage Device's StorageController
		//			If VirtualizationManagement:
		//				Allocate Resources

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

	model.Controllers.OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StorageControllers", storageId)

	return nil
}

// StorageIdStoragePoolsGet -
func StorageIdStoragePoolsGet(storageId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StoragePools/%s", storageId, defaultStoragePoolId)

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
		model.Members[idx].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StorageControllers/%s", storageId, c.id)
	}

	return nil
}

func StorageIdControllerIdGet(storageId, controllerId string, model *sf.StorageControllerV100StorageController) error {

	_, err := findStorageController(storageId, controllerId)
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

func StorageIdVolumesGet(storageId string, model *sf.VolumeCollectionVolumeCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	// TODO: Query the volumes on the device

	return nil
}

func StorageIdVolumePost(storageId string, model *sf.VolumeV161Volume) (string, error) {
	return "", nil
}

func StorageIdVolumeIdGet(storageId, volumeId string, model *sf.VolumeV161Volume) error {
	return nil
}

func StorageIdVolumeIdDelete(storageId, volumeId string) error {
	return nil
}
