package nnf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "stash.us.cray.com/rabsw/nnf-ec/internal/events"

	fabric "stash.us.cray.com/rabsw/nnf-ec/internal/manager-fabric"
	nvme "stash.us.cray.com/rabsw/nnf-ec/internal/manager-nvme"
	server "stash.us.cray.com/rabsw/nnf-ec/internal/manager-server"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"stash.us.cray.com/rabsw/ec"
	openapi "stash.us.cray.com/rabsw/rfsf-openapi/pkg/common"
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

type StorageService struct {
	id string

	config                   *ConfigFile
	serverControllerProvider server.ServerControllerProvider

	pools       []StoragePool
	groups      []StorageGroup
	endpoints   []Endpoint
	fileSystems []FileSystem

	// Index of the Id field of any Storage Service resource (Pools, Groups, Endpoints, FileSystems)
	// That is, given a Storage Service resource OdataId field, ResourceIndex will correspond to the
	// index within the OdataId splity by "/" i.e.     strings.split(OdataId, "/")[ResourceIndex]
	resourceIndex int
}

type StoragePool struct {
	id          string
	name        string
	description string

	uid    uuid.UUID
	policy AllocationPolicy

	allocatedVolume  AllocatedVolume
	providingVolumes []ProvidingVolume

	storageGroups []*StorageGroup
	fileSystem    *FileSystem

	storageService *StorageService
}

type AllocatedVolume struct {
	id            string
	capacityBytes int64

	storagePool *StoragePool // Is this needed?
}

type ProvidingVolume struct {
	volume *nvme.Volume
}

type Endpoint struct {
	id           string
	name         string
	controllerId uint16
	state        sf.ResourceState

	fabricId string

	storageService *StorageService
}

type StorageGroup struct {
	id string

	volume           *AllocatedVolume
	storageEndpoints []StorageGroupEndpoint

	storagePool    *StoragePool
	storageService *StorageService
}

type StorageGroupEndpoint struct {
	// Endpoint represents the server for this storage group; this is analgous to
	// the Endpoints represented in the Fabric Manager of Initiator type.
	endpoint *Endpoint

	// Server represents a connection to the physical server endpoint that manages the
	// storage devices. This can be locally managed on the NNF contorller itself, or
	// remotely managed through some magical being not yet determined.
	serverStorage *server.ServerStoragePool
}

type FileSystem struct {
	id          string
	accessModes []string

	fsApi  server.FileSystemApi
	shares []ExportedFileShare

	storagePool    *StoragePool
	storageService *StorageService
}

type ExportedFileShare struct {
	id        string
	mountRoot string

	sgEndpoint *StorageGroupEndpoint
	fileSystem *FileSystem
}

const (
	DefaultStorageServiceId            = "unassigned" // This is loaded from config
	DefaultStoragePoolCapacitySourceId = "0"
	DefaultAllocatedVolumeId           = "0"
)

var storageService = StorageService{}

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

func findExportedFileShare(storageServiceId, fileSystemId, fileShareId string) (*StorageService, *FileSystem, *ExportedFileShare) {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return nil, nil, nil
	}

	return s, fs, fs.findExportedFileShare(fileShareId)
}

func (s *StorageService) fmt(format string, a ...interface{}) string {
	return fmt.Sprintf("/redfish/v1/StorageServices/%s", s.id) + fmt.Sprintf(format, a...)
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

func (s *StorageService) createStoragePool(uid uuid.UUID, name string, description string, policy AllocationPolicy, providingVolumes []ProvidingVolume) *StoragePool {

	// Find a free Storage Pool Id
	var poolId = -1
	for _, p := range s.pools {
		id, _ := strconv.Atoi(p.id)

		if poolId <= id {
			poolId = id
		}
	}

	poolId = poolId + 1

	capacityBytes := int64(0)
	for _, v := range providingVolumes {
		capacityBytes += int64(v.volume.GetCapaityBytes())
	}

	p := &StoragePool{
		id:               strconv.Itoa(poolId),
		name:             name,
		description:      description,
		uid:              uid,
		policy:           policy,
		providingVolumes: providingVolumes,
		storageService:   s,
	}

	p.allocatedVolume = AllocatedVolume{
		id:            DefaultAllocatedVolumeId,
		capacityBytes: capacityBytes,
		storagePool:   p,
	}

	if p.name == "" {
		p.name = fmt.Sprintf("Storage Pool %s", p.id)
	}

	return p
}

func (s *StorageService) createStorageGroup(volume *AllocatedVolume, endpoints []*Endpoint) *StorageGroup {

	// Find a free Storage Group Id
	var groupId = -1
	for _, g := range s.groups {
		id, _ := strconv.Atoi(g.id)

		if groupId <= id {
			groupId = id
		}
	}

	groupId = groupId + 1

	isNnf := func(ep *Endpoint) bool {
		model := sf.EndpointV150Endpoint{}
		fabric.FabricIdEndpointsEndpointIdGet(ep.fabricId, ep.id, &model)
		if model.ConnectedEntities[0].EntityType == sf.PROCESSOR_EV150ET {
			return true
		}

		return false
	}

	storageEndpoints := make([]StorageGroupEndpoint, len(endpoints))
	for idx, ep := range endpoints {
		ctrl := s.serverControllerProvider.NewServerController(
			server.ServerControllerOptions{Local: isNnf(ep)})

		storageEndpoints[idx] = StorageGroupEndpoint{
			endpoint:      ep,
			serverStorage: ctrl.NewServerStoragePool(volume.storagePool.uid),
		}
	}

	return &StorageGroup{
		id:               strconv.Itoa(groupId),
		volume:           volume,
		storageEndpoints: storageEndpoints,
		storagePool:      volume.storagePool,
		storageService:   s,
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

func (s *StorageService) createFileSystem(sp *StoragePool, fsApi server.FileSystemApi) *FileSystem {

	// Find a free File System Id
	var fileSystemId = -1
	for _, fs := range s.fileSystems {
		id, _ := strconv.Atoi(fs.id)

		if fileSystemId <= id {
			fileSystemId = id
		}
	}

	fileSystemId = fileSystemId + 1

	return &FileSystem{
		id:             strconv.Itoa(fileSystemId),
		fsApi:          fsApi,
		storagePool:    sp,
		storageService: s,
	}
}

func (p *StoragePool) fmt(format string, a ...interface{}) string {
	return p.storageService.fmt("/StoragePools/%s", p.id) + fmt.Sprintf(format, a...)
}

func (p *StoragePool) isCapacitySource(capacitySourceId string) bool {
	return capacitySourceId == DefaultStoragePoolCapacitySourceId
}

func (p *StoragePool) isAllocatedVolume(volumeId string) bool {
	return volumeId == DefaultAllocatedVolumeId
}

func (p *StoragePool) capacitySourcesGet() []sf.CapacityCapacitySource {
	return []sf.CapacityCapacitySource{
		{
			OdataId:   p.fmt("/CapacitySources"),
			OdataType: "#CapacitySource.v1_0_0.CapacitySource",
			Name:      "Capacity Source",
			Id:        DefaultStoragePoolCapacitySourceId,

			ProvidedCapacity: sf.CapacityV120Capacity{
				// TODO
			},

			ProvidingVolumes: sf.OdataV4IdRef{OdataId: p.fmt("/CapacitySources/%s/ProvidingVolumes", DefaultStoragePoolCapacitySourceId)},
		},
	}
}

func (p *StoragePool) findStorageGroupEndpoint(endpoint *Endpoint) *StorageGroupEndpoint {
	for _, sg := range p.storageGroups {
		for idx, sge := range sg.storageEndpoints {
			if sge.endpoint == endpoint {
				return &sg.storageEndpoints[idx]
			}
		}
	}

	return nil
}

func (ep *Endpoint) fmt(format string, a ...interface{}) string {
	return ep.storageService.fmt("/Endpoints/%s", ep.id) + fmt.Sprintf(format, a...)
}

func (sg *StorageGroup) fmt(format string, a ...interface{}) string {
	return sg.storageService.fmt("/StorageGroups/%s", sg.id) + fmt.Sprintf(format, a...)
}

func (sg *StorageGroup) status() sf.ResourceStatus {
	state := sf.UNAVAILABLE_OFFLINE_RST
	for _, sge := range sg.storageEndpoints {
		state = getResourceStateFromServerStoragePoolStatus(sge.serverStorage.GetStatus())
		if state != sf.ENABLED_RST {
			break
		}
	}

	return sf.ResourceStatus{
		Health: sf.OK_RH,
		State:  state,
	}
}

func (fs *FileSystem) fmt(format string, a ...interface{}) string {
	return fs.storageService.fmt("/FileSystems/%s", fs.id) + fmt.Sprintf(format, a...)
}

func (fs *FileSystem) findExportedFileShare(id string) *ExportedFileShare {
	for fileShareIdx, fileShare := range fs.shares {
		if fileShare.id == id {
			return &fs.shares[fileShareIdx]
		}
	}

	return nil
}

func (fs *FileSystem) createFileShare(sgEndpoint *StorageGroupEndpoint, mountRoot string) *ExportedFileShare {
	var fileShareId = -1
	for _, fileShare := range fs.shares {
		id, _ := strconv.Atoi(fileShare.id)

		if fileShareId <= id {
			fileShareId = id
		}
	}

	fileShareId = fileShareId + 1

	return &ExportedFileShare{
		id:         strconv.Itoa(fileShareId),
		sgEndpoint: sgEndpoint,
		mountRoot:  mountRoot,
		fileSystem: fs,
	}
}

func (sh *ExportedFileShare) initialize(mountpoint string) error {
	return sh.sgEndpoint.serverStorage.CreateFileSystem(sh.fileSystem.fsApi, mountpoint)
}

func (sh *ExportedFileShare) getStatus() sf.ResourceStatus {

	return sf.ResourceStatus{
		Health: sf.OK_RH,
		State:  getResourceStateFromServerStoragePoolStatus(sh.sgEndpoint.serverStorage.GetStatus()),
	}
}

func (sh *ExportedFileShare) fmt(format string, a ...interface{}) string {
	return sh.fileSystem.fmt("/ExportedFileShares/%s", sh.id) + fmt.Sprintf(format, a...)
}

func Initialize(ctrl NnfControllerInterface) error {

	storageService = StorageService{
		id:                       DefaultStorageServiceId,
		serverControllerProvider: ctrl.ServerControllerProvider(),
	}

	s := &storageService

	conf, err := loadConfig()
	if err != nil {
		log.WithError(err).Errorf("Failed to load %s configuration", s.id)
		return err
	}

	log.Debugf("%+v", conf)

	s.id = conf.Id
	s.config = conf

	log.Debugf("NNF Storage Service '%s' Loaded...", conf.Metadata.Name)
	log.Debugf("  Server Config    : %+v", conf.ServerConfig)
	log.Debugf("  Allocation Config: %+v", conf.AllocationConfig)

	s.endpoints = make([]Endpoint, conf.ServerConfig.Count)
	for endpointIdx := range s.endpoints {
		s.endpoints[endpointIdx] = Endpoint{
			id:             strconv.Itoa(endpointIdx),
			state:          sf.UNAVAILABLE_OFFLINE_RST,
			storageService: s,
		}
	}

	s.resourceIndex = strings.Count(s.fmt("/StoragePool/0"), "/")

	PortEventManager.Subscribe(PortEventSubscriber{
		HandlerFunc: PortEventHandler,
		Data:        s,
	})

	// Initialize the Server Manager - considered internal to
	// the NNF Manager
	if err := server.Initialize(); err != nil {
		log.WithError(err).Errorf("Failed to Initialize Server Manager")
		return err
	}

	return nil
}

func PortEventHandler(event PortEvent, data interface{}) {
	s := data.(*StorageService)

	if event.PortType != PORT_TYPE_USP {
		return
	}

	ep, err := fabric.GetEndpointFromPortEvent(event)
	if err != nil {
		log.WithError(err).Errorf("Unable to find endpoint for event %+v", event)
		return
	}

	endpointIdx := ep.Index()
	if !(endpointIdx < len(s.endpoints)) {
		log.Errorf("Endpoint index %d beyond supported endpoint count %d", endpointIdx, len(s.endpoints))
		return
	}

	endpoint := Endpoint{
		id:             ep.Id(),
		name:           ep.Name(),
		fabricId:       event.FabricId,
		controllerId:   ep.ControllerId(),
		state:          sf.ENABLED_RST, // TODO: Port Down Ev
		storageService: s,
	}

	s.endpoints[endpointIdx] = endpoint
}

func Get(model *sf.StorageServiceCollectionStorageServiceCollection) error {

	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0] = sf.OdataV4IdRef{
		OdataId: fmt.Sprintf("/redfish/v1/StorageServices/%s", storageService.id),
	}

	return nil
}

func StorageServiceIdGet(storageServiceId string, model *sf.StorageServiceV150StorageService) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.Id = s.id
	model.StoragePools = sf.OdataV4IdRef{OdataId: s.fmt("/StoragePools")}
	model.StorageGroups = sf.OdataV4IdRef{OdataId: s.fmt("/StorageGroups")}
	model.Endpoints = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints")}
	model.FileSystems = sf.OdataV4IdRef{OdataId: s.fmt("/FileSystems")}

	return nil
}

func StorageServiceIdStoragePoolsGet(storageServiceId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(s.pools))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for poolIdx, pool := range s.pools {
		model.Members[poolIdx] = sf.OdataV4IdRef{OdataId: s.fmt("/StoragePools/%s", pool.id)}
	}

	return nil
}

// StorageServiceIdStoragePoolsPost -
func StorageServiceIdStoragePoolsPost(storageServiceId string, model *sf.StoragePoolV150StoragePool) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrBadRequest
	}

	// TODO: Check the model for valid RAID configurations

	policy := NewAllocationPolicy(s.config.AllocationConfig, model.Oem)
	if policy == nil {
		log.Errorf("Failed to allocate storage policy. Config: %+v Oem: %+v", s.config.AllocationConfig, model.Oem)
		return ec.ErrNotAcceptable
	}

	capacityBytes := model.CapacityBytes
	if capacityBytes == 0 {
		capacityBytes = model.Capacity.Data.AllocatedBytes
	}

	if capacityBytes == 0 {
		return ec.ErrNotAcceptable
	}

	if err := policy.Initialize(uint64(capacityBytes)); err != nil {
		log.WithError(err).Errorf("Failed to initialize storage policy")
		return ec.ErrInternalServerError
	}

	if err := policy.CheckCapacity(); err != nil {
		log.WithError(err).Warnf("Storage Policy does not provide sufficient capacity to support requested %d bytes", capacityBytes)
		return ec.ErrNotAcceptable
	}

	// All checks have completed; we're ready to create the storage pool
	uid := s.allocateStoragePoolUid()

	log.Infof("Allocating storage for PID %s", uid.String())
	volumes, err := policy.Allocate(uid)
	if err != nil {

		// TOOD: there may be partial volumes for this pool - we need to mark
		// them for delete in the ledger and attempt to release them.

		log.WithError(err).Errorf("Storage Policy allocation failed.")
		return ec.ErrInternalServerError
	}

	p := s.createStoragePool(uid, model.Name, model.Description, policy, volumes)
	s.pools = append(s.pools, *p)

	return StorageServiceIdStoragePoolIdGet(storageServiceId, p.id, model)
}

// StorageServiceIdStoragePoolIdGet -
func StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrBadRequest
	}

	model.Id = p.id
	model.OdataId = p.fmt("")
	model.AllocatedVolumes = sf.OdataV4IdRef{OdataId: p.fmt("/AllocatedVolumes")}

	model.BlockSizeBytes = 4096 // TODO
	model.Capacity = sf.CapacityV100Capacity{
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes:   p.allocatedVolume.capacityBytes,
			ProvisionedBytes: p.allocatedVolume.capacityBytes,
			ConsumedBytes:    0, // TODO
		},
	}

	model.CapacityBytes = p.allocatedVolume.capacityBytes
	model.CapacitySources = p.capacitySourcesGet()
	model.CapacitySourcesodataCount = int64(len(model.CapacitySources))

	model.Identifier = sf.ResourceIdentifier{
		DurableName:       p.uid.String(),
		DurableNameFormat: sf.UUID_RV1100DNF,
	}

	model.Links.StorageGroupsodataCount = int64(len(p.storageGroups))
	model.Links.StorageGroups = make([]sf.OdataV4IdRef, model.Links.StorageGroupsodataCount)
	for idx, sg := range p.storageGroups {
		model.Links.StorageGroups[idx] = sf.OdataV4IdRef{OdataId: sg.fmt("")}
	}

	if p.fileSystem != nil {
		model.Links.FileSystem = sf.OdataV4IdRef{OdataId: p.fileSystem.fmt("")}
	}

	return nil
}

// StorageServiceIdStoragePoolIdDelete -
func StorageServiceIdStoragePoolIdDelete(storageServiceId, storagePoolId string) error {
	s, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	for _, pv := range p.providingVolumes {
		if err := nvme.DeleteVolume(pv.volume); err != nil {
			log.WithError(err).Errorf("Failed to delete volume from storage pool %s", p.id)
			return ec.ErrInternalServerError
		}

		// TODO: If any delete fails, we're left with dangling volumes preventing
		// further deletion. Need to fix or recover from this. Maybe a transaction
		// log.
	}

	for idx, pool := range s.pools {
		if pool.id == storagePoolId {
			copy(s.pools[idx:], s.pools[idx+1:])
			s.pools = s.pools[:len(s.pools)-1]
		}
	}

	return nil
}

// StorageServiceIdStoragePoolIdCapacitySourcesGet -
func StorageServiceIdStoragePoolIdCapacitySourcesGet(storageServiceId, storagePoolId string, model *sf.CapacitySourceCollectionCapacitySourceCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	model.Members = p.capacitySourcesGet()
	model.MembersodataCount = int64(len(model.Members))

	return nil
}

// StorageServiceIdStoragePoolIdCapacitySourceIdGet -
func StorageServiceIdStoragePoolIdCapacitySourceIdGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.CapacityCapacitySource) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	if !p.isCapacitySource(capacitySourceId) {
		return ec.ErrNotFound
	}

	s := p.capacitySourcesGet()[0]

	model.Id = s.Id
	model.ProvidedCapacity = s.ProvidedCapacity
	model.ProvidingVolumes = s.ProvidingVolumes

	return nil
}

// StorageServiceIdStoragePoolIdCapacitySourceIdProvidingVolumesGet -
func StorageServiceIdStoragePoolIdCapacitySourceIdProvidingVolumesGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.VolumeCollectionVolumeCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	if !p.isCapacitySource(capacitySourceId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(p.providingVolumes))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, v := range p.providingVolumes {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: v.volume.GetOdataId()}
	}

	return nil
}

// StorageServiceIdStoragePoolIdAlloctedVolumesGet -
func StorageServiceIdStoragePoolIdAlloctedVolumesGet(storageServiceId, storagePoolId string, model *sf.VolumeCollectionVolumeCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = 1
	model.Members = []sf.OdataV4IdRef{
		{OdataId: p.fmt("/AllocatedVolumes/%s", DefaultAllocatedVolumeId)},
	}

	return nil
}

// StorageServiceIdStoragePoolIdAllocatedVolumeIdGet -
func StorageServiceIdStoragePoolIdAllocatedVolumeIdGet(storageServiceId, storagePoolId, volumeId string, model *sf.VolumeV161Volume) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	if !p.isAllocatedVolume(volumeId) {
		return ec.ErrNotFound
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
func StorageServiceIdStorageGroupsGet(storageServiceId string, model *sf.StorageGroupCollectionStorageGroupCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(s.groups))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for groupIdx, group := range s.groups {
		model.Members[groupIdx] = sf.OdataV4IdRef{OdataId: s.fmt("StorageGroups/%s", group.id)}
	}

	return nil
}

// StorageServiceIdStorageGroupPost -
func StorageServiceIdStorageGroupPost(storageServiceId string, model *sf.StorageGroupV150StorageGroup) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	if model.MappedVolumes == nil || len(model.MappedVolumes) != 1 {
		return ec.ErrBadRequest
	}

	// This is the expected format of the Volume Odata Id...
	// "/redfish/v1/StorageServices/NNF/StoragePools/0/AllocatedVolumes/0"
	fmt := s.fmt("/StoragePools/%s/AllocatedVolumes/%s", "(?P<storagePoolId>\\w+)", "(?P<volumeId>\\w+)")

	re := regexp.MustCompile(fmt)
	matches := re.FindStringSubmatch(model.MappedVolumes[0].Volume.OdataId)
	if matches == nil {
		return ec.ErrBadRequest
	}

	storagePoolId := matches[re.SubexpIndex("storagePoolId")]
	volumeId := matches[re.SubexpIndex("volumeId")]

	_, sp := findStoragePool(storageServiceId, storagePoolId)
	if sp == nil {
		return ec.ErrBadRequest
	}

	if volumeId != DefaultAllocatedVolumeId {
		return ec.ErrBadRequest
	}

	// Ensure the provided server endpoints are valid
	if model.ServerEndpoints == nil || len(model.ServerEndpoints) == 0 {
		return ec.ErrBadRequest
	}

	endpoints := []*Endpoint{}
	for _, ep := range model.ServerEndpoints {
		fields := strings.Split(ep.OdataId, "/")
		if len(fields) != s.resourceIndex+1 {
			return ec.ErrBadRequest
		}

		endpointId := fields[s.resourceIndex]

		e := s.findEndpoint(endpointId)
		if e == nil {
			return ec.ErrBadRequest
		}

		if e.state != sf.ENABLED_RST {
			return ec.ErrInternalServerError
		}

		endpoints = append(endpoints, e)
	}

	// Everything validated OK - create the Storage Group

	sg := s.createStorageGroup(&sp.allocatedVolume, endpoints)
	sp.storageGroups = append(sp.storageGroups, sg)
	s.groups = append(s.groups, *sg)

	model.Id = sg.id
	model.OdataId = sg.fmt("")

	return StorageServiceIdStorageGroupIdGet(storageServiceId, sg.id, model)
}

// StorageServiceIdStorageGroupIdGet -
func StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	s, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	// TODO: Mapped Volumes should point to the corresponding Storage Volume
	//       As they are present on the Server Storage Controller.

	// TODO:
	// model.Volumes - Should point to the volumes which are present on the Storage Endpoint,
	//                 once exposed. This means iterating on the Storage Endpoints and asking
	//                 every Storage Controller for status on the volumes in the Storage Pool.
	model.MappedVolumes = []sf.StorageGroupMappedVolume{
		{
			AccessCapability: sf.READ_WRITE_SGAC,
			Volume:           sf.OdataV4IdRef{OdataId: sg.volume.storagePool.fmt("/AllocatedVolume/%s", sg.volume.id)},
		},
	}

	// This is a little odd since we only ever support one Server per
	// Storage Group. Maybe refactor the rfsf implementation for this
	// to be singular.
	model.ServerEndpointsodataCount = int64(len(sg.storageEndpoints))
	model.ServerEndpoints = make([]sf.OdataV4IdRef, model.ServerEndpointsodataCount)
	for seIdx, se := range sg.storageEndpoints {
		model.ServerEndpoints[seIdx] = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints/%s", se.endpoint.id)}
	}

	model.Links.StoragePool = sf.OdataV4IdRef{OdataId: sg.storagePool.fmt("")}

	model.Status = sg.status()

	return nil
}

// StorageServiceIdStorageGroupIdDelete -
func StorageServiceIdStorageGroupIdDelete(storageServiceId, storageGroupId string) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	// TODO: This should result in a bunch of detach of the volume
	//       from it's controllers. The storage group is then deleted
	//       but the storage pool still exists.

	return nil
}

// StorageServiceIdStorageGroupIdExposeVolumesPost -
func StorageServiceIdStorageGroupIdExposeVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150ExposeVolumes) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	sp := sg.volume.storagePool

	controllerIds := make([]uint16, len(sg.storageEndpoints))
	for seIdx, se := range sg.storageEndpoints {
		controllerIds[seIdx] = se.endpoint.controllerId

	}

	for _, volume := range sp.providingVolumes {
		if err := nvme.AttachControllers(volume.volume, controllerIds); err != nil {
			// TODO: Rollback
			return err
		}
	}

	return nil
}

// StorageServiceIdStorageGroupIdHideVolumesPost -
func StorageServiceIdStorageGroupIdHideVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150HideVolumes) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	return ec.ErrInternalServerError
}

// StorageServiceIdEndpointsGet -
func StorageServiceIdEndpointsGet(storageServiceId string, model *sf.EndpointCollectionEndpointCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(s.endpoints))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, ep := range s.endpoints {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints/%s", ep.id)}
	}

	return nil
}

// StorageServiceIdEndpointIdGet -
func StorageServiceIdEndpointIdGet(storageServiceId, endpointId string, model *sf.EndpointV150Endpoint) error {
	_, e := findEndpoint(storageServiceId, endpointId)
	if e == nil {
		return ec.ErrNotFound
	}

	if err := fabric.FabricIdEndpointsEndpointIdGet(e.fabricId, e.id, model); err != nil {
		return err
	}

	return nil
}

// StorageServiceIdFileSystemsGet -
func StorageServiceIdFileSystemsGet(storageServiceId string, model *sf.FileSystemCollectionFileSystemCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(s.fileSystems))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, fileSystem := range s.fileSystems {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: s.fmt("/FileSystems/%s", fileSystem.id)}
	}

	return nil
}

// StorageServiceIdFileSystemsPost -
func StorageServiceIdFileSystemsPost(storageServiceId string, model *sf.FileSystemV122FileSystem) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	// Extract the StoragePoolId from the POST model
	fields := strings.Split(model.StoragePool.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.ErrNotAcceptable
	}
	storagePoolId := fields[s.resourceIndex]

	// Find the existing storage pool - the file system will link to the providing pool
	sp := s.findStoragePool(storagePoolId)
	if sp == nil {
		return ec.ErrNotAcceptable
	}

	if sp.fileSystem != nil {
		return ec.ErrNotAcceptable
	}

	// TODO: Oem should define the file system type (by name)
	type Oem struct {
		Type string
	}

	oem := Oem{}

	if err := openapi.UnmarshalOem(model.Oem, &oem); err != nil {
		return ec.ErrBadRequest
	}

	fsApi := server.FileSystemController.NewFileSystem(oem.Type)
	if fsApi == nil {
		return ec.ErrNotAcceptable
	}

	fs := s.createFileSystem(sp, fsApi)
	sp.fileSystem = fs
	s.fileSystems = append(s.fileSystems, *fs)

	return StorageServiceIdFileSystemIdGet(storageServiceId, fs.id, model)
}

// StorageServiceIdFileSystemIdGet -
func StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	model.Id = fs.id
	model.OdataId = fs.fmt("")

	model.CapacityBytes = fs.storagePool.allocatedVolume.capacityBytes
	model.StoragePool = sf.OdataV4IdRef{OdataId: s.fmt("/StoragePools/%s", fs.storagePool.id)}
	model.ExportedShares = sf.OdataV4IdRef{OdataId: fs.fmt("/ExportedFileShares")}

	return nil
}

// StorageServiceIdFileSystemIdDelete -
func StorageServiceIdFileSystemIdDelete(storageServiceId, fileSystemId string) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	// Require all shares to be unmounted prior deletion
	// TODO: Or do it automatically?
	if len(fs.shares) != 0 {
		return ec.ErrNotAcceptable
	}

	for fileSystemIdx, fileSystem := range s.fileSystems {
		if fileSystem.id == fileSystemId {
			s.fileSystems = append(s.fileSystems[:fileSystemIdx], s.fileSystems[fileSystemIdx+1:]...)
			break
		}
	}

	return nil
}

// StorageServiceIdFileSystemIdExportedSharesGet -
func StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId string, model *sf.FileShareCollectionFileShareCollection) error {
	_, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fs.shares))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, sh := range fs.shares {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: sh.fmt("")}
	}
	return nil
}

// StorageServiceIdFileSystemIdExportedSharesPost -
func StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId string, model *sf.FileShareV120FileShare) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	fields := strings.Split(model.ServerEndpoint.OdataId, "/")
	if len(fields) != s.resourceIndex+1 {
		return ec.ErrNotAcceptable
	}

	endpointId := fields[s.resourceIndex]
	ep := s.findEndpoint(endpointId)
	if ep == nil {
		return ec.ErrNotAcceptable
	}

	if len(model.FileSharePath) == 0 {
		return ec.ErrNotAcceptable
	}

	// Find the Storage Group Endpoint - There should be a Storage Group
	// Endpoint that has an association to the fs.storagePool and endpoint.
	// This represents the physical devices on the server that backs the
	// File System and supports the Exported Share.
	sge := fs.storagePool.findStorageGroupEndpoint(ep)
	if sge == nil {
		return ec.ErrNotAcceptable
	}

	sh := fs.createFileShare(sge, model.FileSharePath)
	fs.shares = append(fs.shares, *sh)

	// TODO: Until the share is ready, the file system will no exist. We should add a timeout
	// to this request - wait for the share to come ready and then create the file system.
	for sh.getStatus().State == sf.ENABLED_RST {
		time.Sleep(time.Second)
	}

	if err := sh.initialize(model.FileSharePath); err != nil {
		log.WithError(err).Errorf("Failed to initialize file share for path %s", model.FileSharePath)
	}

	model.Id = sh.id
	model.OdataId = sh.fmt("")

	return StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, sh.id, model)
}

// StorageServiceIdFileSystemIdExportedShareIdGet -
func StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) error {
	_, fs, sh := findExportedFileShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.ErrNotFound
	}

	model.Id = sh.id
	model.OdataId = sh.fmt("")
	model.FileSharePath = sh.mountRoot
	// model.ServerEndpoint = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints/%s", sh.endpoint.id)} // TODO: Remove this from model
	model.Links.FileSystem = sf.OdataV4IdRef{OdataId: fs.fmt("")}
	model.Links.Endpoint = sf.OdataV4IdRef{OdataId: sh.sgEndpoint.endpoint.fmt("")}

	model.Status = sh.getStatus() // TODO

	return nil
}

func StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId string) error {
	_, fs, sh := findExportedFileShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.ErrNotFound
	}

	for shareIdx, share := range fs.shares {
		if share.id == exportedShareId {
			fs.shares = append(fs.shares[:shareIdx], fs.shares[shareIdx+1:]...)
			break
		}
	}

	// TODO: Actually need to unmount the share

	return nil
}

func getResourceStateFromServerStoragePoolStatus(status server.ServerStoragePoolStatus) sf.ResourceState {
	switch status {
	case server.ServerStoragePoolStarting:
		return sf.STARTING_RST
	case server.ServerStoragePoolReady:
		return sf.ENABLED_RST
	}
	return sf.DISABLED_RST
}
