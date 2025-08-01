/*
 * Copyright 2020-2025 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nnf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	ec "github.com/NearNodeFlash/nnf-ec/pkg/ec"
	event "github.com/NearNodeFlash/nnf-ec/pkg/manager-event"
	fabric "github.com/NearNodeFlash/nnf-ec/pkg/manager-fabric"
	msgreg "github.com/NearNodeFlash/nnf-ec/pkg/manager-message-registry/registries"
	nvme "github.com/NearNodeFlash/nnf-ec/pkg/manager-nvme"
	server "github.com/NearNodeFlash/nnf-ec/pkg/manager-server"
	"github.com/NearNodeFlash/nnf-ec/pkg/persistent"
	openapi "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/common"
	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
)

// Logging Keys used for named arguments. Should be kept in lowerCamelCase per k8s standard.
// https://github.com/kubernetes/community/blob/HEAD/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments
const (
	modelIdKey        = "id"
	storagePoolIdKey  = "storagePoolId"
	storageGroupIdKey = "storageGroupId"
	fileSystemIdKey   = "fileSystemId"
	fileShareIdKey    = "fileShareId"
	endpointIdKey     = "endpointId"
	odataIdKey        = "odataId"
)

var storageService = StorageService{
	id:     DefaultStorageServiceId,
	state:  sf.DISABLED_RST,
	health: sf.CRITICAL_RH,
}

func NewDefaultStorageService(unknownVolumes bool, replaceMissingVolumes bool) StorageServiceApi {
	storageService.deleteUnknownVolumes = unknownVolumes
	storageService.replaceMissingVolumes = replaceMissingVolumes
	return NewAerService(&storageService) // Wrap the default storage service with Advanced Error Reporting capabilities
}

type StorageService struct {
	id     string
	state  sf.ResourceState
	health sf.ResourceHealth

	config                   *ConfigFile
	store                    *persistent.Store
	serverControllerProvider server.ServerControllerProvider
	persistentController     PersistentControllerApi

	pools       []StoragePool
	groups      []StorageGroup
	endpoints   []Endpoint
	fileSystems []FileSystem

	// Index of the Id field of any Storage Service resource (Pools, Groups, Endpoints, FileSystems)
	// That is, given a Storage Service resource OdataId field, ResourceIndex will correspond to the
	// index within the OdataId split by "/" i.e.     strings.split(OdataId, "/")[ResourceIndex]
	resourceIndex int

	// This flag controls whether we delete volumes that don't appear in storage pools we know about.
	deleteUnknownVolumes bool
	// This flag controls whether we replace volumes that are missing from storage pools.
	replaceMissingVolumes bool

	log ec.Logger
}

func (s *StorageService) OdataId() string {
	return fmt.Sprintf("/redfish/v1/StorageServices/%s", s.id)
}

func (s *StorageService) OdataIdRef(ref string) sf.OdataV4IdRef {
	return sf.OdataV4IdRef{OdataId: fmt.Sprintf("%s%s", s.OdataId(), ref)}
}

func (s *StorageService) GetStore() *persistent.Store {
	return s.store
}

func (s *StorageService) findStoragePool(storagePoolId string) *StoragePool {
	for poolIdx, pool := range s.pools {
		if pool.id == storagePoolId {
			return &s.pools[poolIdx]
		}
	}

	return nil
}

func (s *StorageService) findStorageGroup(storageGroupId string) *StorageGroup {
	for groupIdx, group := range s.groups {
		if group.id == storageGroupId {
			return &s.groups[groupIdx]
		}
	}

	return nil
}

func (s *StorageService) findEndpoint(endpointId string) *Endpoint {
	for endpointIdx, endpoint := range s.endpoints {
		if endpoint.id == endpointId {
			return &s.endpoints[endpointIdx]
		}
	}

	return nil
}

func (s *StorageService) findFileSystem(fileSystemId string) *FileSystem {
	for fileSystemIdx, fileSystem := range s.fileSystems {
		if fileSystem.id == fileSystemId {
			return &s.fileSystems[fileSystemIdx]
		}
	}

	return nil
}

// Create a storage pool object with the provided variables and add it to the storage service's list of storage
// pools. If an ID is not provided, an unused one will be used. If an ID is provided, the caller must check
// that the ID does not already exist.
func (s *StorageService) createStoragePool(id, name, description string, uid uuid.UUID, policy AllocationPolicy) *StoragePool {

	// If no ID is supplied, find a free Storage Pool Id
	if len(id) == 0 {
		var poolId = -1
		for _, p := range s.pools {
			if id, err := strconv.Atoi(p.id); err == nil {
				if poolId <= id {
					poolId = id
				}
			}
		}

		poolId = poolId + 1
		id = strconv.Itoa(poolId)
	}

	if uid.Variant() == uuid.Invalid {
		uid = s.allocateStoragePoolUid()
	}

	s.pools = append(s.pools, StoragePool{
		id:             id,
		name:           name,
		description:    description,
		uid:            uid,
		policy:         policy,
		storageService: s,
	})

	return &s.pools[len(s.pools)-1]
}

func (s *StorageService) deleteStoragePool(sp *StoragePool) {

	for storagePoolIdx, storagePool := range s.pools {
		if storagePool.id == sp.id {
			s.pools = append(s.pools[:storagePoolIdx], s.pools[storagePoolIdx+1:]...)
			break
		}
	}
}

func (s *StorageService) patchStoragePool(sp *StoragePool, forceRescan bool) error {
	log := s.log

	// In a test environment where you may be deleting volumes underneath nnf-ec and
	// the initial rescan has already completed, it is handly to force a rescan
	if !forceRescan && len(sp.missingVolumes) == 0 {
		log.V(2).Info("No missing volumes to replace")
		return nil
	}

	// Look for missing volumes
	err := sp.checkVolumes()
	if err != nil {
		log.Error(err, "Unable to rescan volumes")
		return err
	}

	err = sp.replaceMissingVolumes()
	if err != nil {
		log.Error(err, "Unable to replace missing volumes")
		return err
	}

	// Persist the changes
	updateFunc := func() error {
		// Nothing to do for simple metadata updates
		return nil
	}

	if err := s.persistentController.UpdatePersistentObject(sp, updateFunc, storagePoolStorageUpdateStartLogEntryType, storagePoolStorageUpdateCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithError(err).WithCause("Failed to update storage pool")
	}

	return err
}

func (s *StorageService) findStorage(sn string) *nvme.Storage {
	for _, storage := range nvme.GetStorage() {
		if storage.SerialNumber() == sn {
			return storage
		}
	}

	return nil
}

// Create a storage group object with the provided variables and add it to the storage service's list of storage
// groups. If an ID is not provided, an unused one will be used. If an ID is provided, the caller must check
// that the ID does not already exist.
func (s *StorageService) createStorageGroup(id string, sp *StoragePool, endpoint *Endpoint) *StorageGroup {

	if len(id) == 0 {
		// Find a free Storage Group Id
		var groupId = -1
		for _, sg := range s.groups {
			if id, err := strconv.Atoi(sg.id); err == nil {

				if groupId <= id {
					groupId = id
				}
			}
		}

		groupId = groupId + 1
		id = strconv.Itoa(groupId)
	}

	expectedNamespaces := make([]server.StorageNamespace, len(sp.providingVolumes))
	for idx, pv := range sp.providingVolumes {
		volume := pv.Storage.FindVolume(pv.VolumeId)
		if volume == nil {
			err := fmt.Errorf("Volume not found")
			s.log.Error(err, "Storage pool createStorageGroup volume not found", "volumeid", pv.VolumeId)
			continue
		}

		expectedNamespaces[idx] = server.StorageNamespace{
			SerialNumber: pv.Storage.SerialNumber(),
			Id:           volume.GetNamespaceId(),
		}
	}

	s.groups = append(s.groups, StorageGroup{
		id:             id,
		endpoint:       endpoint,
		serverStorage:  endpoint.serverCtrl.NewStorage(sp.uid, expectedNamespaces),
		storagePoolId:  sp.id,
		storageService: s,
	})

	sg := &s.groups[len(s.groups)-1]

	sp.storageGroupIds = append(sp.storageGroupIds, id)

	return sg
}

func (s *StorageService) deleteStorageGroup(sg *StorageGroup) {
	sp := s.findStoragePool(sg.storagePoolId)

	for storageGroupIdx, storageGroupId := range sp.storageGroupIds {
		if storageGroupId == sg.id {
			sp.storageGroupIds = append(sp.storageGroupIds[:storageGroupIdx], sp.storageGroupIds[storageGroupIdx+1:]...)
			break
		}
	}

	for storageGroupIdx, storageGroup := range s.groups {
		if storageGroup.id == sg.id {
			s.groups = append(s.groups[:storageGroupIdx], s.groups[storageGroupIdx+1:]...)
			break
		}
	}
}

func (s *StorageService) allocateStoragePoolUid() uuid.UUID {
	for {
	Retry:
		uid := uuid.New()

		for _, p := range s.pools {
			if p.uid == uid {
				goto Retry
			}
		}

		return uid
	}
}

// Create a file system object with the provided variables and add it to the storage service's list of file
// systems. If an ID is not provided, an unused one will be used. If an ID is provided, the caller must check
// that the ID does not already exist.
func (s *StorageService) createFileSystem(id string, sp *StoragePool, fsApi server.FileSystemApi, fsOem server.FileSystemOem) *FileSystem {

	if len(id) == 0 {
		// Find a free File System Id
		var fileSystemId = -1
		for _, fs := range s.fileSystems {
			if id, err := strconv.Atoi(fs.id); err == nil {
				if fileSystemId <= id {
					fileSystemId = id
				}
			}
		}

		fileSystemId = fileSystemId + 1
		id = strconv.Itoa(fileSystemId)
	}

	sp.fileSystemId = id

	s.fileSystems = append(s.fileSystems, FileSystem{
		id:             id,
		fsApi:          fsApi,
		fsOem:          fsOem,
		storagePoolId:  sp.id,
		storageService: s,
	})

	return &s.fileSystems[len(s.fileSystems)-1]
}

func (s *StorageService) deleteFileSystem(fs *FileSystem) {
	sp := s.findStoragePool(fs.storagePoolId)

	sp.fileSystemId = ""

	for fileSystemIdx, fileSystem := range s.fileSystems {
		if fileSystem.id == fs.id {
			s.fileSystems = append(s.fileSystems[:fileSystemIdx], s.fileSystems[fileSystemIdx+1:]...)
			break
		}
	}
}

const (
	DefaultCapacitySourceId            = "0"
	DefaultStorageServiceId            = "unassigned" // This is loaded from config
	DefaultStoragePoolCapacitySourceId = "0"
	DefaultAllocatedVolumeId           = "0"
)

func isStorageService(storageServiceId string) bool { return storageServiceId == storageService.id }
func findStorageService(storageServiceId string) *StorageService {
	if !isStorageService(storageServiceId) {
		return nil
	}

	return &storageService
}

func findStoragePool(storageServiceId, storagePoolId string) (*StorageService, *StoragePool) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return nil, nil
	}

	return s, s.findStoragePool(storagePoolId)
}

func findStorageGroup(storageServiceId, storageGroupId string) (*StorageService, *StorageGroup) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return nil, nil
	}

	return s, s.findStorageGroup(storageGroupId)
}

func findEndpoint(storageServiecId, endpointId string) (*StorageService, *Endpoint) {
	s := findStorageService(storageServiecId)
	if s == nil {
		return nil, nil
	}

	return s, s.findEndpoint(endpointId)
}

func findFileSystem(storageServiceId, fileSystemId string) (*StorageService, *FileSystem) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return nil, nil
	}

	return s, s.findFileSystem(fileSystemId)
}

func findFileShare(storageServiceId, fileSystemId, fileShareId string) (*StorageService, *FileSystem, *FileShare) {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return nil, nil, nil
	}

	return s, fs, fs.findFileShare(fileShareId)
}

func (s *StorageService) Id() string {
	return s.id
}

func (s *StorageService) cleanupVolumes() {
	// Build a list of all providing volumes from all storage pools
	var providingVolumes []nvme.ProvidingVolume
	for _, pool := range s.pools {
		for _, volume := range pool.providingVolumes {
			providingVolumes = append(providingVolumes, nvme.ProvidingVolume{
				Storage:  volume.Storage,
				VolumeId: volume.VolumeId,
			})
		}
	}

	nvme.CleanupVolumes(providingVolumes)
}

// Initialize is responsible for initializing the NNF Storage Service; the
// Storage Service must complete initialization without error prior to any
// access to the Storage Service. Failure to initialize will cause the
// storage service to misbehave.
func (s *StorageService) Initialize(log ec.Logger, ctrl NnfControllerInterface) error {

	log.Info("Initialize NNF Manager")

	storageService.state = sf.STARTING_RST

	// Dynamic controllers for managing objects. These are typically programmed via command line arguments
	// the the NNF Controller Interface is created
	storageService.serverControllerProvider = ctrl.ServerControllerProvider()
	storageService.persistentController = ctrl.PersistentControllerProvider()

	// Reserve space for the most common allocation types. 32 is the current
	// limit for the number of supported namespaces.
	storageService.pools = make([]StoragePool, 0, 32)
	storageService.groups = make([]StorageGroup, 0, 32)
	storageService.fileSystems = make([]FileSystem, 0, 32)

	const name = "nnf"
	log.V(2).Info("Creating logger", "name", name)
	log = log.WithName(name)
	storageService.log = log

	conf, err := loadConfig()
	if err != nil {
		log.Error(err, "failed to load configuration", "id", s.id)
		return err
	}

	s.id = conf.Id
	s.config = conf

	s.endpoints = make([]Endpoint, len(conf.RemoteConfig.Servers))
	for endpointIdx := range s.endpoints {
		opts := server.ServerControllerOptions{
			Local:   true,
			Address: "",
		}
		s.endpoints[endpointIdx] = Endpoint{
			id:             strconv.Itoa(endpointIdx),
			state:          sf.UNAVAILABLE_OFFLINE_RST,
			config:         &conf.RemoteConfig.Servers[endpointIdx],
			storageService: s,
			serverCtrl:     s.serverControllerProvider.NewServerController(opts),
		}
	}

	// Index of a individual resource located off of the collections managed
	// by the NNF Storage Service.
	s.resourceIndex = strings.Count(s.OdataIdRef("/StoragePool/0").OdataId, "/")

	// Create the key-value storage database
	{
		path := "nnf.db"
		persistent.SetLogger(log)
		s.store, err = persistent.Open(path, false)
		if err != nil {
			log.Error(err, "Unable to open database", "path", path)
			return err
		}

		s.store.Register([]persistent.Registry{
			NewStoragePoolRecoveryRegistry(s),
			NewStorageGroupRecoveryRegistry(s),
			NewFileSystemRecoveryRegistry(s),
			NewFileShareRecoveryRegistry(s),
		})
	}

	// Initialize the Server Manager - considered internal to
	// the NNF Manager
	if err := server.Initialize(); err != nil {
		log.Error(err, "Failed to Initialize Server Manager")
		return err
	}

	// Subscribe ourselves to events
	event.EventManager.Subscribe(s)

	return nil
}

func (s *StorageService) Close() error {
	return s.store.Close()
}

func (s *StorageService) EventHandler(e event.Event) error {
	log := s.log.WithValues("eventId", e.Id, "eventMessage", e.Message, "eventArgs", e.MessageArgs)

	// Upstream link events
	linkEstablished := e.Is(msgreg.UpstreamLinkEstablishedFabric("", "")) || e.Is(msgreg.DegradedUpstreamLinkEstablishedFabric("", ""))
	linkDropped := e.Is(msgreg.UpstreamLinkDroppedFabric("", ""))

	if linkEstablished || linkDropped {
		log.V(2).Info("Link event")

		var switchId, portId string
		if err := e.Args(&switchId, &portId); err != nil {
			return ec.NewErrInternalServerError().WithError(err).WithCause("event parameters illformed")
		}

		ep, err := fabric.GetEndpoint(switchId, portId)
		if err != nil {
			return ec.NewErrInternalServerError().WithError(err).WithCause("failed to locate endpoint")
		}

		endpoint := &s.endpoints[ep.Index()]

		endpoint.id = ep.Id()
		endpoint.name = ep.Name()
		endpoint.controllerId = ep.ControllerId()
		endpoint.fabricId = fabric.FabricId

		if linkEstablished {
			log.V(2).Info("Link established")
			opts := server.ServerControllerOptions{
				Local:   ep.Type() == sf.PROCESSOR_EV150ET,
				Address: endpoint.config.Address,
			}

			endpoint.serverCtrl = s.serverControllerProvider.NewServerController(opts)

			if endpoint.serverCtrl.Connected() {
				endpoint.state = sf.ENABLED_RST
			} else {
				endpoint.state = sf.STANDBY_OFFLINE_RST
			}

		} else if linkDropped {
			log.Info("Link dropped")
			endpoint.state = sf.UNAVAILABLE_OFFLINE_RST
		}

		return nil
	}

	// Fabric is ready
	// All devices are enumerated and discovery is complete.
	if e.Is(msgreg.FabricReadyNnf("")) {
		log.V(1).Info("Fabric ready")

		if err := s.store.Replay(); err != nil {
			log.Error(err, "Failed to replay storage database")
			return err
		}

		// Remove any namespaces that are not part of a Storage Pool
		if s.deleteUnknownVolumes {
			log.V(2).Info("Cleanup unknown volumes")
			s.cleanupVolumes()
		}

		if s.replaceMissingVolumes {
			log.V(2).Info("Replace missing volumes")
			for spIdx := range s.pools {
				pool := &s.pools[spIdx]
				if err := pool.storageService.patchStoragePool(pool, false /* rescan */); err != nil {
					log.Error(err, "Failed to replace missing volumes", "poolId", pool.id)
				}
			}
		}

		s.state = sf.ENABLED_RST
		s.health = sf.OK_RH

		var fabricID string
		if err := e.Args(&fabricID); err != nil {
			return ec.NewErrInternalServerError().WithError(err).WithCause("event parameters illformed")
		}

		f := &sf.FabricV120Fabric{}
		if err := fabric.FabricIdGet(fabricID, f); err != nil {
			return ec.NewErrInternalServerError().WithError(err).WithCause("fabric not found")
		}

		if s.health == sf.OK_RH {
			s.health = f.Status.Health
		}

		log.Info("Storage Service Enabled", "health", s.health)

		nvme.StartNVMeMonitor(s.log)
	}

	// Check for storage pool events
	if e.Is(msgreg.StoragePoolPatchedNnf("", "", "", "", "")) {
		// After a storage pool is patched, check for new volumes that need to be attached
		log.V(1).Info("Storage Pool Patched")
		var storagePoolID string
		var oldStorageSN string
		var oldNamespaceID string
		var newStorageSN string
		var newNamespaceID string
		if err := e.Args(&storagePoolID, &oldStorageSN, &oldNamespaceID, &newStorageSN, &newNamespaceID); err != nil {
			return ec.NewErrInternalServerError().WithError(err).WithCause("event parameters illformed")
		}

		log = log.WithValues("poolId", storagePoolID, "oldStorageSN", oldStorageSN, "oldNamespaceId", oldNamespaceID, "newStorageSN", newStorageSN, "newNamespaceId", newNamespaceID)

		// We may have multiple storage groups associated with the same storage pool
		for _, sg := range s.groups {
			if sg.storagePoolId == storagePoolID {
				if err := sg.recoverPool(); err != nil {
					return ec.NewErrInternalServerError().WithError(err).WithCause("unable to update storage group")
				}
			}
		}
	}

	return nil
}

//
// Methods below this comment block are all service handlers that provide the various
// CRUD (Create, Read, Update, and Delete) functionality. The method pattern is always
// the same:
//    StorageService[Sub Resource ...][HTTP Method]([Ids string...], model *[Swordfish Model])
// where
//    Sub Resource: defines the hierarchy of resources, beginning at the NNF Storage
//       Service and into the Redfish/Swordfish object tree
//    HTTP Method: defines a HTTP access method. One of...
//       Post (Create): Used to create an object in the Storage Service
//       Get (Read): Used to read an object from the Storage Service
//       Patch (Update): Used to update an existing object managed by the Storage Service
//       Delete: Used to delete an object managed by the Storage Service
//    IDs: defines one or more IDs that identify a resource
//    Swordfish Model: a Swordfish structure for which HTTP Method should act on.
//       Create: the method will attempt to create the object from the provided model
//       Get: the method will populate the model.
//       Patch: the method will update parameters with parameters from the model
//       Delete: There is no model parameter.
//

func (*StorageService) StorageServicesGet(model *sf.StorageServiceCollectionStorageServiceCollection) error {

	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0] = sf.OdataV4IdRef{
		OdataId: fmt.Sprintf("/redfish/v1/StorageServices/%s", storageService.id),
	}

	return nil
}

func (*StorageService) StorageServiceIdGet(storageServiceId string, model *sf.StorageServiceV150StorageService) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.Id = s.id

	model.Status.State = s.state
	model.Status.Health = s.health

	model.StoragePools = s.OdataIdRef("/StoragePools")
	model.StorageGroups = s.OdataIdRef("/StorageGroups")
	model.Endpoints = s.OdataIdRef("/Endpoints")
	model.FileSystems = s.OdataIdRef("/FileSystems")

	model.Links.CapacitySource = s.OdataIdRef("/CapacitySource")
	return nil
}

func (*StorageService) StorageServiceIdCapacitySourceGet(storageServiceId string, model *sf.CapacityCapacitySource) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.Id = DefaultCapacitySourceId

	totalCapacityBytes, totalUnallocatedBytes := uint64(0), uint64(0)
	if err := nvme.EnumerateStorage(func(odataId string, capacityBytes uint64, unallocatedBytes uint64) {

		// TODO: OdataId could be used to link to the underlying StoragePool that is the NVMe Device
		//       That would require a new Storage Pool Collection that lives off the CapacitySource and
		//       is properly linked.
		totalCapacityBytes += capacityBytes
		totalUnallocatedBytes += unallocatedBytes
	}); err != nil {
		return ec.NewErrInternalServerError().WithError(err).WithCause("Failed to enumerate storage")
	}

	model.ProvidedCapacity.Data.GuaranteedBytes = int64(totalUnallocatedBytes)
	model.ProvidedCapacity.Data.ProvisionedBytes = int64(totalCapacityBytes)
	model.ProvidedCapacity.Data.AllocatedBytes = int64(totalCapacityBytes - totalUnallocatedBytes)
	model.ProvidedCapacity.Data.ConsumedBytes = model.ProvidedCapacity.Data.AllocatedBytes

	return nil
}

func (*StorageService) StorageServiceIdStoragePoolsGet(storageServiceId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.MembersodataCount = int64(len(s.pools))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for poolIdx, pool := range s.pools {
		model.Members[poolIdx] = sf.OdataV4IdRef{OdataId: pool.OdataId()}
	}

	return nil
}

// StorageServiceIdStoragePoolsPost - create a storage pool
func (*StorageService) StorageServiceIdStoragePoolsPost(storageServiceId string, model *sf.StoragePoolV150StoragePool) (err error) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	log := s.log.WithValues(modelIdKey, model.Id)
	log.V(2).Info("Creating storage pool")
	defer func() {
		if err != nil {
			log.Error(err, "Create storage pool failed")
		}
	}()

	policy := NewAllocationPolicy(s.config.AllocationConfig, model.Oem)
	if policy == nil {
		return ec.NewErrNotAcceptable().WithEvent(msgreg.PropertyValueTypeErrorBase("Oem", fmt.Sprintf("%+v", model.Oem)))
	}

	capacityInBytes := model.CapacityBytes
	if capacityInBytes == 0 {
		capacityInBytes = model.Capacity.Data.AllocatedBytes
	}

	if capacityInBytes == 0 {
		return ec.NewErrNotAcceptable().WithEvent(msgreg.CreateFailedMissingReqPropertiesBase("CapacityBytes"))
	}

	if err := policy.Initialize(uint64(capacityInBytes)); err != nil {
		log.Error(err, "Failed to initialize storage policy", "capacityInBytes", capacityInBytes)
		return ec.NewErrInternalServerError().WithResourceType(StorageServiceOdataType).WithError(err).WithCause("Failed to initialize storage policy")
	}

	if err := policy.CheckAndAdjustCapacity(); err != nil {
		log.Error(err, "Storage policy cannot support capacity", "capacityInBytes", capacityInBytes)
		return ec.NewErrNotAcceptable().WithResourceType(StorageServiceOdataType).WithError(err).WithCause("Insufficient capacity available")
	}

	p := s.createStoragePool(model.Id, model.Name, model.Description, uuid.UUID{}, policy)

	updateFunc := func() (err error) {
		p.providingVolumes, err = policy.Allocate()
		if err != nil {
			return err
		}

		p.allocatedVolume = AllocatedVolume{
			id:            DefaultAllocatedVolumeId,
			capacityBytes: p.GetCapacityBytes(),
		}

		return nil
	}

	if err := s.persistentController.CreatePersistentObject(p, updateFunc, storagePoolStorageCreateStartLogEntryType, storagePoolStorageCreateCompleteLogEntryType); err != nil {
		s.deleteStoragePool(p)
		return ec.NewErrInternalServerError().WithResourceType(StorageServiceOdataType).WithError(err).WithCause("Failed to allocate storage volumes")
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceCreatedResourceEvent(), p)

	log.Info("Created storage pool", storagePoolIdKey, p.id, "volumes", len(p.providingVolumes), "capacityInBytes", p.allocatedVolume.capacityBytes)

	return s.StorageServiceIdStoragePoolIdGet(storageServiceId, p.id, model)
}

// StorageServiceIdStoragePoolsPatch updates the storage pools in the storage service.
func (*StorageService) StorageServiceIdStoragePoolsPatch(storageServiceId string, model *sf.StoragePoolCollectionStoragePoolCollection) (err error) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	log := s.log
	log.V(2).Info("Patching storage pools")
	defer func() {
		if err != nil {
			log.Error(err, "Patch storage pool failed")
		}
	}()

	// Patch each storage pool
	for _, sp := range s.pools {
		log := log.WithValues(storagePoolIdKey, sp.id)
		log.V(2).Info("Patching storage pool")

		poolModel := &sf.StoragePoolV150StoragePool{
			Id: sp.OdataId(),
		}
		err = s.StorageServiceIdStoragePoolIdPatch(storageServiceId, sp.id, poolModel)
		if err != nil {
			break
		}
	}

	model.MembersodataCount = int64(len(s.pools))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for poolIdx, pool := range s.pools {
		model.Members[poolIdx] = sf.OdataV4IdRef{OdataId: pool.OdataId()}
	}

	return nil
}

// StorageServiceIdStoragePoolIdPut -
func (*StorageService) StorageServiceIdStoragePoolIdPut(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	s, p := findStoragePool(storageServiceId, storagePoolId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}
	if p != nil {
		return s.StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId, model)
	}

	model.Id = storagePoolId

	return s.StorageServiceIdStoragePoolsPost(storageServiceId, model)
}

// StorageServiceIdStoragePoolIdGet -
func (*StorageService) StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	s, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	model.Id = p.id
	model.OdataId = p.OdataId()
	model.AllocatedVolumes = p.OdataIdRef("/AllocatedVolumes")

	model.BlockSizeBytes = 4096 // TODO
	model.Capacity = sf.CapacityV100Capacity{
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes:   int64(p.allocatedVolume.capacityBytes),
			ProvisionedBytes: int64(p.allocatedVolume.capacityBytes),
			ConsumedBytes:    0, // TODO
		},
	}

	model.CapacityBytes = int64(p.allocatedVolume.capacityBytes)
	model.CapacitySources = p.capacitySourcesGet()
	model.CapacitySourcesodataCount = int64(len(model.CapacitySources))

	model.Identifier = sf.ResourceIdentifier{
		DurableName:       p.uid.String(),
		DurableNameFormat: sf.UUID_RV1100DNF,
	}

	model.Status.State = sf.ENABLED_RST
	model.Status.Health = sf.OK_RH

	model.Links.StorageGroupsodataCount = int64(len(p.storageGroupIds))
	model.Links.StorageGroups = make([]sf.OdataV4IdRef, model.Links.StorageGroupsodataCount)
	for storageGroupIdx, storageGroupId := range p.storageGroupIds {
		sg := s.findStorageGroup(storageGroupId)
		model.Links.StorageGroups[storageGroupIdx] = sf.OdataV4IdRef{OdataId: sg.OdataId()}
	}

	if p.fileSystemId != "" {
		fs := s.findFileSystem(p.fileSystemId)
		model.Links.FileSystem = sf.OdataV4IdRef{OdataId: fs.OdataId()}
	}

	return nil
}

// StorageServiceIdStoragePoolIdDelete -
func (*StorageService) StorageServiceIdStoragePoolIdDelete(storageServiceId, storagePoolId string) (err error) {
	s, p := findStoragePool(storageServiceId, storagePoolId)

	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	log := s.log.WithValues(storagePoolIdKey, p.id)
	log.V(2).Info("Deleting storage pool")
	defer func() {
		if err != nil {
			log.Error(err, "Delete storage pool failed")
		}
	}()

	if p.fileSystemId != "" {
		if err := s.StorageServiceIdFileSystemIdDelete(s.id, p.fileSystemId); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithError(err).WithCause(fmt.Sprintf("Failed to delete file system '%s'", p.fileSystemId))
		}

		if len(p.fileSystemId) != 0 {
			return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithCause(fmt.Sprintf("File system '%s' not removed from storage pool", p.fileSystemId))
		}
	}

	// Make a copy of the storage groups to be deleted; We can't trust the storage pool's list as
	// it is modified in place within the delete logic
	storageGroupIds := make([]string, len(p.storageGroupIds))
	copy(storageGroupIds, p.storageGroupIds)

	for _, storageGroupId := range storageGroupIds {
		if err := s.StorageServiceIdStorageGroupIdDelete(s.id, storageGroupId); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithError(err).WithCause(fmt.Sprintf("Failed to delete storage group '%s'", storageGroupId))
		}
	}

	if len(p.storageGroupIds) != 0 {
		return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithCause(fmt.Sprintf("Storage groups not removed from storage pool"))
	}

	deleteFunc := func() error {
		err := p.deallocateVolumes()
		if err != nil {
			log.Error(err, "deallocateVolumes failed, but returning success anyway")
		}

		return nil
	}

	if err := s.persistentController.DeletePersistentObject(p, deleteFunc, storagePoolStorageDeleteStartLogEntryType, storagePoolStorageDeleteCompleteLogEntryType); err != nil {
		err := ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithError(err).WithCause(fmt.Sprintf("Failed to delete storage pool"))
		if err != nil {
			log.Error(err, "DeletePersistentObject failed, but returning success anyway")
		}

		return nil
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceRemovedResourceEvent(), p)

	s.deleteStoragePool(p)

	log.Info("Deleted storage pool")

	return nil
}

// StorageServiceIdStoragePoolIdPatch -
func (*StorageService) StorageServiceIdStoragePoolIdPatch(storageServiceID, storagePoolID string, model *sf.StoragePoolV150StoragePool) (err error) {
	s, p := findStoragePool(storageServiceID, storagePoolID)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolID))
	}

	log := s.log.WithValues(storagePoolIdKey, p.id)
	log.V(2).Info("Patching storage pool")
	defer func() {
		if err != nil {
			log.Error(err, "Patch storage pool failed")
		}
	}()

	// Update fields that are allowed to be modified
	if model.Name != "" {
		p.name = model.Name
	}

	if model.Description != "" {
		p.description = model.Description
	}

	// Replace any missing volumes
	if err = s.patchStoragePool(p, true /* forceRescan */); err != nil {
		log.Error(err, "Failed to check and replace volumes in storage pool")
		return ec.NewErrInternalServerError().WithResourceType(StoragePoolOdataType).WithError(err).WithCause("Failed to update storage pool resources")
	}

	log.Info("Patched storage pool")

	// Return the updated storage pool model
	return s.StorageServiceIdStoragePoolIdGet(storageServiceID, storagePoolID, model)
}

// StorageServiceIdStoragePoolIdCapacitySourcesGet -
func (*StorageService) StorageServiceIdStoragePoolIdCapacitySourcesGet(storageServiceId, storagePoolId string, model *sf.CapacitySourceCollectionCapacitySourceCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	model.Members = p.capacitySourcesGet()
	model.MembersodataCount = int64(len(model.Members))

	return nil
}

// StorageServiceIdStoragePoolIdCapacitySourceIdGet -
func (*StorageService) StorageServiceIdStoragePoolIdCapacitySourceIdGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.CapacityCapacitySource) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	if !p.isCapacitySource(capacitySourceId) {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(CapacitySourceOdataType, capacitySourceId))
	}

	s := p.capacitySourcesGet()[0]

	model.Id = s.Id
	model.ProvidedCapacity = s.ProvidedCapacity
	model.ProvidingVolumes = s.ProvidingVolumes

	return nil
}

// StorageServiceIdStoragePoolIdCapacitySourceIdProvidingVolumesGet -
func (*StorageService) StorageServiceIdStoragePoolIdCapacitySourceIdProvidingVolumesGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.VolumeCollectionVolumeCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	if !p.isCapacitySource(capacitySourceId) {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(CapacitySourceOdataType, capacitySourceId))
	}

	model.MembersodataCount = int64(len(p.providingVolumes))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, pv := range p.providingVolumes {
		volume := pv.Storage.FindVolume(pv.VolumeId)
		if volume != nil {
			model.Members[idx] = sf.OdataV4IdRef{OdataId: volume.GetOdataId()}
		}
	}

	return nil
}

// StorageServiceIdStoragePoolIdAllocatedVolumesGet -
func (*StorageService) StorageServiceIdStoragePoolIdAllocatedVolumesGet(storageServiceId, storagePoolId string, model *sf.VolumeCollectionVolumeCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	model.MembersodataCount = 1
	model.Members = []sf.OdataV4IdRef{
		p.OdataIdRef(fmt.Sprintf("/AllocatedVolumes/%s", DefaultAllocatedVolumeId)),
	}

	return nil
}

// StorageServiceIdStoragePoolIdAllocatedVolumeIdGet -
func (*StorageService) StorageServiceIdStoragePoolIdAllocatedVolumeIdGet(storageServiceId, storagePoolId, volumeId string, model *sf.VolumeV161Volume) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	if !p.isAllocatedVolume(volumeId) {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(VolumeOdataType, volumeId))
	}

	model.Id = DefaultAllocatedVolumeId
	model.CapacityBytes = int64(p.allocatedVolume.capacityBytes)
	model.Capacity = sf.CapacityV100Capacity{
		// TODO???
	}

	model.Identifiers = []sf.ResourceIdentifier{
		{
			DurableName:       p.uid.String(),
			DurableNameFormat: sf.NGUID_RV1100DNF,
		},
	}

	model.VolumeType = sf.RAW_DEVICE_VVT

	return nil
}

// StorageServiceIdStorageGroupsGet -
func (*StorageService) StorageServiceIdStorageGroupsGet(storageServiceId string, model *sf.StorageGroupCollectionStorageGroupCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.MembersodataCount = int64(len(s.groups))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for groupIdx, group := range s.groups {
		model.Members[groupIdx] = sf.OdataV4IdRef{OdataId: group.OdataId()}
	}

	return nil
}

// StorageServiceIdStorageGroupPost creates a new storage group in the storage service.
func (*StorageService) StorageServiceIdStorageGroupPost(storageServiceId string, model *sf.StorageGroupV150StorageGroup) (err error) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	log := s.log.WithValues(modelIdKey, model.Id)
	log.V(2).Info("Creating storage group")
	defer func() {
		if err != nil {
			log.Error(err, "Create storage group failed")
		}
	}()

	fields := strings.Split(model.Links.StoragePool.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.InvalidURIBase(model.Links.StoragePool.OdataId))
	}

	storagePoolID := fields[s.resourceIndex]

	_, sp := findStoragePool(storageServiceId, storagePoolID)
	if sp == nil {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolID))
	}

	// TODO RABSW-1110: Ensure storage pool is operational before creating a storage group

	fields = strings.Split(model.Links.ServerEndpoint.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.NewErrNotAcceptable().WithResourceType(EndpointOdataType).WithEvent(msgreg.InvalidURIBase(model.Links.ServerEndpoint.OdataId))
	}

	endpointID := fields[s.resourceIndex]

	ep := s.findEndpoint(endpointID)
	if ep == nil {
		return ec.NewErrNotAcceptable().WithResourceType(EndpointOdataType).WithEvent(msgreg.ResourceNotFoundBase(EndpointOdataType, endpointID))
	}

	if !ep.serverCtrl.Connected() {
		return ec.NewErrNotAcceptable().WithResourceType(EndpointOdataType).WithCause(fmt.Sprintf("Server endpoint '%s' not connected", endpointID))
	}

	// Everything validated OK - create the Storage Group
	sg := s.createStorageGroup(model.Id, sp, ep)

	updateFunc := func() error {
		for _, pv := range sp.providingVolumes {
			volume := pv.Storage.FindVolume(pv.VolumeId)
			if volume == nil {
				return ec.NewErrInternalServerError().WithResourceType(StorageGroupOdataType).WithCause(fmt.Sprintf("Storage group '%s' attach volume '%s' not found", sg.id, pv.VolumeId))
			}

			if err := volume.AttachController(sg.endpoint.controllerId); err != nil {
				return ec.NewErrInternalServerError().WithResourceType(StorageGroupOdataType).WithError(err).WithCause(fmt.Sprintf("Storage group '%s' attach volume '%s' failed", sg.id, pv.VolumeId))
			}
		}

		return nil
	}

	if err := s.persistentController.CreatePersistentObject(sg, updateFunc, storageGroupCreateStartLogEntryType, storageGroupCreateCompleteLogEntryType); err != nil {
		s.deleteStorageGroup(sg)
		return ec.NewErrInternalServerError().WithResourceType(StorageGroupOdataType).WithError(err).WithCause("failed to create storage group")
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceCreatedResourceEvent(), sg)

	log.Info("Created storage group", storageGroupIdKey, sg.id)

	return s.StorageServiceIdStorageGroupIdGet(storageServiceId, sg.id, model)
}

// StorageServiceIdStorageGroupIdPut handles PUT requests for a specific storage group
func (*StorageService) StorageServiceIdStorageGroupIdPut(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	s, sg := findStorageGroup(storageServiceId, storageGroupId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}
	if sg != nil {
		return s.StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId, model)
	}

	model.Id = storageGroupId

	return s.StorageServiceIdStorageGroupPost(storageServiceId, model)
}

// StorageServiceIdStorageGroupIdGet handles GET requests for a specific storage group
func (*StorageService) StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	s, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageGroupOdataType, storageGroupId))
	}

	sp := s.findStoragePool(sg.storagePoolId)
	if sp == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("Storage group '%s' does not have associated storage pool '%s'", storageGroupId, sg.storagePoolId))
	}

	model.Id = sg.id
	model.OdataId = sg.OdataId()

	// TODO: Mapped Volumes should point to the corresponding Storage Volume
	//       As they are present on the Server Storage Controller.

	// TODO:
	// model.Volumes - Should point to the volumes which are present on the Storage Endpoint,
	//                 once exposed. This means iterating on the Storage Endpoints and asking
	//                 every Storage Controller for status on the volumes in the Storage Pool.
	model.MappedVolumes = []sf.StorageGroupMappedVolume{
		{
			AccessCapability: sf.READ_WRITE_SGAC,
			Volume:           sp.OdataIdRef(fmt.Sprintf("/AllocatedVolumes/%s", DefaultAllocatedVolumeId)),
		},
	}

	model.Links.ServerEndpoint = sf.OdataV4IdRef{OdataId: sg.endpoint.OdataId()}
	model.Links.StoragePool = sf.OdataV4IdRef{OdataId: sp.OdataId()}

	model.Status = sg.status()

	return nil
}

// StorageServiceIdStorageGroupIdDelete -
func (*StorageService) StorageServiceIdStorageGroupIdDelete(storageServiceId, storageGroupId string) (err error) {
	s, sg := findStorageGroup(storageServiceId, storageGroupId)

	if sg == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageGroupOdataType, storageGroupId))
	}

	log := s.log.WithValues(storageGroupIdKey, sg.id)
	log.V(2).Info("Deleting storage group")
	defer func() {
		if err != nil {
			log.Error(err, "Delete storage group failed")
		}
	}()

	if sg.fileShareId != "" {
		log.WithValues(fileShareIdKey, sg.fileShareId).Info("cannot delete storage group with existing file share")
		return ec.NewErrNotAcceptable().WithResourceType(StorageGroupOdataType).WithEvent(msgreg.ResourceCannotBeDeletedBase()).WithCause(fmt.Sprintf("Storage group '%s' file share present", storageGroupId))
	}

	sp := s.findStoragePool(sg.storagePoolId)
	if sp == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("Storage group '%s' does not have associated storage pool '%s'", storageGroupId, sg.storagePoolId))
	}

	deleteFunc := func() error {
		// Detach the endpoint from the NVMe namespaces
		for _, pv := range sp.providingVolumes {
			volume := pv.Storage.FindVolume(pv.VolumeId)
			if volume == nil {
				err := fmt.Errorf("Volume not found")
				log.Error(err, "Storage group detach volume not found", "storageGroup", storageGroupId, "volumeid", pv.VolumeId)
				continue
			}

			if err := volume.DetachController(sg.endpoint.controllerId); err != nil {
				log.Error(err, "Storage group failed to detach controller", "storageGroup", storageGroupId, "controller", sg.endpoint.controllerId)
				continue
			}
		}

		// Notify the Server the namespaces were removed
		if err := sg.serverStorage.Delete(); err != nil {
			log.Error(err, "Storage group server delete failed", "storageGroup", storageGroupId)
		}

		return nil
	}

	if err := s.persistentController.DeletePersistentObject(sg, deleteFunc, storageGroupDeleteStartLogEntryType, storageGroupDeleteCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithResourceType(StorageGroupOdataType).WithError(err).WithCause("Failed to delete storage group")
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceRemovedResourceEvent(), sg)

	s.deleteStorageGroup(sg)

	log.Info("Deleted storage group")

	return nil
}

// StorageServiceIdEndpointsGet -
func (*StorageService) StorageServiceIdEndpointsGet(storageServiceId string, model *sf.EndpointCollectionEndpointCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.MembersodataCount = int64(len(s.endpoints))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, ep := range s.endpoints {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: ep.OdataId()}
	}

	return nil
}

// StorageServiceIdEndpointIdGet -
func (*StorageService) StorageServiceIdEndpointIdGet(storageServiceId, endpointId string, model *sf.EndpointV150Endpoint) error {
	_, ep := findEndpoint(storageServiceId, endpointId)
	if ep == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(EndpointOdataType, endpointId))
	}

	model.Id = ep.id
	model.OdataId = ep.OdataId()

	// Ask the fabric manager to fill it the endpoint details
	if err := fabric.FabricIdEndpointsEndpointIdGet(ep.fabricId, ep.id, model); err != nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(EndpointOdataType, ep.id))
	}

	model.OdataId = ep.OdataId() // Done twice so the fabric manager doesn't hijak the @odata.id

	serverInfo := ep.serverCtrl.GetServerInfo()
	model.Oem["LNetNids"] = serverInfo.LNetNids

	return nil
}

// StorageServiceIdFileSystemsGet -
func (*StorageService) StorageServiceIdFileSystemsGet(storageServiceId string, model *sf.FileSystemCollectionFileSystemCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	model.MembersodataCount = int64(len(s.fileSystems))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, fileSystem := range s.fileSystems {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: fileSystem.OdataId()}
	}

	return nil
}

// StorageServiceIdFileSystemsPost -
func (*StorageService) StorageServiceIdFileSystemsPost(storageServiceId string, model *sf.FileSystemV122FileSystem) (err error) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}

	log := s.log.WithValues(modelIdKey, model.Id)
	log.V(2).Info("Create file system")
	defer func() {
		if err != nil {
			log.Error(err, "Create file system failed")
		}
	}()

	// Extract the StoragePoolId from the POST model
	fields := strings.Split(model.Links.StoragePool.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.InvalidURIBase(model.Links.StoragePool.OdataId))
	}
	storagePoolId := fields[s.resourceIndex]

	// Find the existing storage pool - the file system will link to the providing pool
	sp := s.findStoragePool(storagePoolId)
	if sp == nil {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.ResourceNotFoundBase(StoragePoolOdataType, storagePoolId))
	}

	if sp.fileSystemId != "" {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithCause(fmt.Sprintf("Storage pool '%s' no file system defined", storagePoolId))
	}

	oem := server.FileSystemOem{}
	if err := openapi.UnmarshalOem(model.Oem, &oem); err != nil {
		return ec.NewErrBadRequest().WithResourceType(FileSystemOdataType).WithError(err).WithEvent(msgreg.MalformedJSONBase())
	}

	fsApi, err := server.FileSystemController.NewFileSystem(&oem)
	if err != nil {
		return ec.NewErrNotAcceptable().WithResourceType(FileSystemOdataType).WithError(err).WithCause("File system '%s' failed to allocate").WithEvent(msgreg.InternalErrorBase())
	}
	if fsApi == nil {
		return ec.NewErrNotAcceptable().WithResourceType(FileSystemOdataType).WithEvent(msgreg.PropertyValueNotInListBase(oem.Type, "Type"))
	}

	fs := s.createFileSystem(model.Id, sp, fsApi, oem)

	if err := s.persistentController.CreatePersistentObject(fs, func() error { return nil }, fileSystemCreateStartLogEntryType, fileSystemCreateCompleteLogEntryType); err != nil {
		s.deleteFileSystem(fs)
		return ec.NewErrInternalServerError().WithResourceType(FileSystemOdataType).WithError(err).WithCause(fmt.Sprintf("File system '%s' failed to create", fs.id))
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceCreatedResourceEvent(), fs)

	log.Info("Created file system", fileSystemIdKey, fs.id)

	return s.StorageServiceIdFileSystemIdGet(storageServiceId, fs.id, model)
}

// StorageServiceIdFileSystemIdPut -
func (*StorageService) StorageServiceIdFileSystemIdPut(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if s == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(StorageServiceOdataType, storageServiceId))
	}
	if fs != nil {
		return s.StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId, model)
	}

	model.Id = fileSystemId

	return s.StorageServiceIdFileSystemsPost(storageServiceId, model)
}

// StorageServiceIdFileSystemIdGet -
func (*StorageService) StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileSystemOdataType, fileSystemId))
	}

	sp := s.findStoragePool(fs.storagePoolId)
	if sp == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("Could not find storage pool for file system Storage Pool ID: %s", fs.storagePoolId))
	}

	model.Id = fs.id
	model.OdataId = fs.OdataId()

	model.CapacityBytes = int64(sp.allocatedVolume.capacityBytes)
	model.StoragePool = sf.OdataV4IdRef{OdataId: sp.OdataId()}
	model.ExportedShares = fs.OdataIdRef("/ExportedFileShares")

	model.Oem = openapi.MarshalOem(fs.fsOem)

	return nil
}

// StorageServiceIdFileSystemIdDelete -
func (*StorageService) StorageServiceIdFileSystemIdDelete(storageServiceId, fileSystemId string) (err error) {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileSystemOdataType, fileSystemId))
	}

	log := s.log.WithValues(fileSystemIdKey, fs.id)
	log.V(2).Info("Deleting file system")
	defer func() {
		if err != nil {
			log.Error(err, "Delete file system failed")
		}
	}()

	// Create a copy of file share IDs; The deletion of a share will modify the fs.shares[] so we cannot
	// iterate on that array directly as it is editted in place.
	shareIds := make([]string, len(fs.shares))
	for idx, sh := range fs.shares {
		shareIds[idx] = sh.id
	}

	for _, shid := range shareIds {
		if err := s.StorageServiceIdFileSystemIdExportedShareIdDelete(s.id, fs.id, shid); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(FileSystemOdataType).WithError(err).WithCause(fmt.Sprintf("Exported share '%s' failed delete", shid))
		}
	}

	if err := s.persistentController.DeletePersistentObject(fs, func() error { return nil }, fileSystemDeleteStartLogEntryType, fileSystemDeleteCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithResourceType(FileSystemOdataType).WithError(err).WithCause("Failed to delete file system")
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceRemovedResourceEvent(), fs)

	s.deleteFileSystem(fs)

	log.Info("Deleted file system")

	return nil
}

// StorageServiceIdFileSystemIdExportedSharesGet -
func (*StorageService) StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId string, model *sf.FileShareCollectionFileShareCollection) error {
	_, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileSystemOdataType, fileSystemId))
	}

	model.MembersodataCount = int64(len(fs.shares))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, sh := range fs.shares {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: sh.OdataId()}
	}
	return nil
}

// StorageServiceIdFileSystemIdExportedSharesPost -
func (*StorageService) StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId string, model *sf.FileShareV120FileShare) (err error) {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileSystemOdataType, fileSystemId))
	}

	log := s.log.WithValues(modelIdKey, model.Id)
	log.V(2).Info("Creating file share")
	defer func() {
		if err != nil {
			log.Error(err, "Create file share failed")
		}
	}()

	fields := strings.Split(model.Links.Endpoint.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.NewErrNotAcceptable().WithResourceType(FileSystemOdataType).WithEvent(msgreg.InvalidURIBase(model.Links.Endpoint.OdataId))
	}

	endpointId := fields[s.resourceIndex]
	ep := s.findEndpoint(endpointId)
	if ep == nil {
		return ec.NewErrNotAcceptable().WithResourceType(EndpointOdataType).WithEvent(msgreg.ResourceNotFoundBase(EndpointOdataType, endpointId))
	}

	sp := s.findStoragePool(fs.storagePoolId)
	if sp == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("Could not find storage pool for file system Storage Pool ID: %s", fs.storagePoolId))
	}

	// Find the Storage Group Endpoint - There should be a Storage Group
	// Endpoint that has an association to the fs.storagePool and endpoint.
	// This represents the physical devices on the server that back the
	// File System and supports the Exported Share.
	sg := sp.findStorageGroupByEndpoint(ep)
	if sg == nil {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.ResourceNotFoundBase(StorageGroupOdataType, endpointId))
	}

	// Wait for the storage group to be ready (enabled) to ensure the disks are present on the system
	state := sg.status().State
	if state == sf.STARTING_RST {
		log.V(2).Info("Storage group starting", storageGroupIdKey, sg.id)
		return ec.NewErrorNotReady().WithResourceType(StorageGroupOdataType).WithCause(fmt.Sprintf("Storage group '%s' is starting", sg.id))
	} else if state != sf.ENABLED_RST {
		log.Info("Storage group not enabled", storageGroupIdKey, sg.id)
		return ec.NewErrNotAcceptable().WithResourceType(StorageGroupOdataType).WithCause(fmt.Sprintf("Storage group '%s' is not ready: state %s", sg.id, state))
	}

	sh := fs.createFileShare(model.Id, sg, model.FileSharePath)

	createFunc := func() error {
		if err := sg.serverStorage.CreateFileSystem(fs.fsApi, model.Oem); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("File share '%s' create failed", sh.id))
		}

		if err := sg.serverStorage.MountFileSystem(fs.fsApi, sh.mountRoot); err != nil {
			if deleteErr := sg.serverStorage.DeleteFileSystem(fs.fsApi); deleteErr != nil {
				return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(deleteErr).WithCause(fmt.Sprintf("File share '%s' failed delete after mount failure", sh.id))
			}

			return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("File share '%s' mount failed", sh.id))
		}

		return nil
	}

	if err := s.persistentController.CreatePersistentObject(sh, createFunc, fileShareCreateStartLogEntryType, fileShareCreateCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithError(err).WithCause(fmt.Sprintf("File share '%s' failed to create", sh.id))
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceCreatedResourceEvent(), sh)

	log.Info("Created file share", fileShareIdKey, sh.id, "path", sh.mountRoot)

	return s.StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, sh.id, model)
}

// StorageServiceIdFileSystemIdExportedShareIdPut -
func (*StorageService) StorageServiceIdFileSystemIdExportedShareIdPut(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) (err error) {
	s, fs, sh := findFileShare(storageServiceId, fileSystemId, exportedShareId)
	if fs == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileShareOdataType, exportedShareId))
	}

	if sh == nil {
		model.Id = exportedShareId
		return s.StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId, model)
	}

	log := s.log.WithValues(fileShareIdKey, sh.id, fileSystemIdKey, fs.id)
	log.V(1).Info("Updating file share")
	defer func() {
		if err != nil {
			log.Error(err, "Update file share failed")
		}
	}()

	newPath := model.FileSharePath
	if err := s.StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId, model); err != nil {
		return err
	}

	if model.FileSharePath == newPath {
		return nil
	}

	fields := strings.Split(model.Links.Endpoint.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.NewErrNotAcceptable().WithResourceType(FileSystemOdataType).WithEvent(msgreg.InvalidURIBase(model.Links.Endpoint.OdataId))
	}

	endpointId := fields[s.resourceIndex]
	ep := s.findEndpoint(endpointId)
	if ep == nil {
		return ec.NewErrNotAcceptable().WithResourceType(EndpointOdataType).WithEvent(msgreg.ResourceNotFoundBase(EndpointOdataType, endpointId))
	}

	sp := s.findStoragePool(fs.storagePoolId)
	if sp == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("Could not find storage pool for file system Storage Pool ID: %s", fs.storagePoolId))
	}

	// Find the Storage Group Endpoint - There should be a Storage Group
	// Endpoint that has an association to the fs.storagePool and endpoint.
	// This represents the physical devices on the server that back the
	// File System and supports the Exported Share.
	sg := sp.findStorageGroupByEndpoint(ep)
	if sg == nil {
		return ec.NewErrNotAcceptable().WithResourceType(StoragePoolOdataType).WithEvent(msgreg.ResourceNotFoundBase(StorageGroupOdataType, endpointId))
	}

	var updateFunc func() error
	if len(newPath) != 0 {
		if len(sh.mountRoot) != 0 {
			// TODO Error: File System already mounted
		}

		sh.mountRoot = newPath

		updateFunc = func() error {
			if err := sg.serverStorage.MountFileSystem(fs.fsApi, sh.mountRoot); err != nil {
				return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("Failed to mount file share '%s' at path '%s'", sh.id, sh.mountRoot))
			}

			return nil
		}

	} else {
		updateFunc = func() error {
			if err := sg.serverStorage.UnmountFileSystem(fs.fsApi, sh.mountRoot); err != nil {
				return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("Failed to unmount file share '%s' at path '%s'", sh.id, sh.mountRoot))
			}

			sh.mountRoot = ""

			return nil
		}
	}

	if err := s.persistentController.UpdatePersistentObject(sh, updateFunc, fileShareUpdateStartLogEntryType, fileShareUpdateCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithError(err).WithCause(fmt.Sprintf("File share '%s' failed to update", sh.id))
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceChangedResourceEvent(), sh)

	log.V(1).Info("Updated file share")

	return s.StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, sh.id, model)

}

// StorageServiceIdFileSystemIdExportedShareIdGet -
func (*StorageService) StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) error {
	s, fs, sh := findFileShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileShareOdataType, exportedShareId))
	}

	sg := s.findStorageGroup(sh.storageGroupId)
	if sg == nil {
		return ec.NewErrInternalServerError().WithCause(fmt.Sprintf("File share '%s' does not have associated storage group '%s'", exportedShareId, sh.storageGroupId))
	}

	model.Id = sh.id
	model.OdataId = sh.OdataId()
	model.FileSharePath = sh.mountRoot
	model.Links.FileSystem = sf.OdataV4IdRef{OdataId: fs.OdataId()}
	model.Links.Endpoint = sf.OdataV4IdRef{OdataId: sg.endpoint.OdataId()}

	model.Status = *sh.getStatus() // TODO

	return nil
}

// StorageServiceIdFileSystemIdExportedShareIdDelete -
func (*StorageService) StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId string) (err error) {
	s, fs, sh := findFileShare(storageServiceId, fileSystemId, exportedShareId)

	if sh == nil {
		return ec.NewErrNotFound().WithEvent(msgreg.ResourceNotFoundBase(FileShareOdataType, exportedShareId))
	}

	log := s.log.WithValues(fileShareIdKey, sh.id)
	log.V(2).Info("Deleting file share")
	defer func() {
		if err != nil {
			log.Error(err, "Delete file share failed")
		}
	}()

	sg := s.findStorageGroup(sh.storageGroupId)
	if sg == nil {
		return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithCause(fmt.Sprintf("File share '%s' does not have associated storage group '%s'", exportedShareId, sh.storageGroupId))
	}

	deleteFunc := func() error {
		if err := sg.serverStorage.UnmountFileSystem(fs.fsApi, sh.mountRoot); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("File share '%s' failed unmount", exportedShareId))
		}

		if err := sg.serverStorage.DeleteFileSystem(fs.fsApi); err != nil {
			return ec.NewErrInternalServerError().WithResourceType(FileShareOdataType).WithError(err).WithCause(fmt.Sprintf("File share '%s' failed delete", exportedShareId))
		}

		return nil
	}

	if err := s.persistentController.DeletePersistentObject(sh, deleteFunc, fileShareDeleteStartLogEntryType, fileShareDeleteCompleteLogEntryType); err != nil {
		return ec.NewErrInternalServerError().WithError(err).WithResourceType(FileShareOdataType).WithCause("Failed to delete file share")
	}

	event.EventManager.PublishResourceEvent(msgreg.ResourceRemovedResourceEvent(), sh)

	fs.deleteFileShare(sh)

	log.Info("Deleted file share")

	return nil
}
