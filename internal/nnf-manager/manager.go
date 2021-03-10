package nnf

import (
	"fmt"
	"strconv"
	"strings"

	"stash.us.cray.com/rabsw/nnf-ec/internal/api"
	. "stash.us.cray.com/rabsw/nnf-ec/internal/events"
	nvme "stash.us.cray.com/rabsw/nnf-ec/internal/nvme-namespace-manager"

	log "github.com/sirupsen/logrus"

	"stash.us.cray.com/rabsw/ec"
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

type StorageService struct {
	id string

	config *ConfigFile

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
	id     string
	policy AllocationPolicy

	allocatedVolume  AllocatedVolume
	providingVolumes []ProvidingVolume

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
	id              string
	controllerIndex uint16
	state           sf.ResourceState

	storageService *StorageService
}

type StorageGroup struct {
	id string

	volume    *AllocatedVolume
	endpoints []*Endpoint

	storageService *StorageService
}

type FileSystem struct {
	id          string
	accessModes []string

	storagePool *StoragePool
	shares      []FileShare

	storageService *StorageService
}

type FileShare struct {
	id        string
	endpoint  *Endpoint
	mountRoot string
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

func findFileSystem(storageServiceId, fileSystemId string) (*StorageService, *FileSystem) {
	s := findStorageService(storageServiceId)
	if s == nil {
		return nil, nil
	}

	return s, s.findFileSystem(fileSystemId)
}

func findExportedShare(storageServiceId, fileSystemId, fileShareId string) (*StorageService, *FileSystem, *FileShare) {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return nil, nil, nil
	}

	return s, fs, fs.findFileShare(fileShareId)
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

func (s *StorageService) createStoragePool(policy AllocationPolicy, providingVolumes []ProvidingVolume) *StoragePool {

	// Find a free Storage Pool Id
	var poolId = -1
	for _, p := range s.pools {
		id, _ := strconv.Atoi(p.id)

		if poolId <= id {
			poolId = id
		}
	}

	poolId = poolId + 1

	p := &StoragePool{
		id:               strconv.Itoa(poolId),
		policy:           policy,
		providingVolumes: providingVolumes,
		storageService:   s,
	}

	capacityBytes := int64(0)
	for _, v := range providingVolumes {
		capacityBytes += int64(v.volume.GetCapaityBytes())
	}

	p.allocatedVolume = AllocatedVolume{
		id:            DefaultAllocatedVolumeId,
		capacityBytes: capacityBytes,
		storagePool:   p,
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

	return &StorageGroup{
		id:             strconv.Itoa(groupId),
		storageService: s,
		volume:         volume,
		endpoints:      endpoints,
	}
}

func (s *StorageService) createFileSystem(sp *StoragePool) *FileSystem {

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

func (sg *StorageGroup) fmt(format string, a ...interface{}) string {
	return sg.storageService.fmt("/StorageGroups/%s", sg.id) + fmt.Sprintf(format, a...)
}

func (fs *FileSystem) fmt(format string, a ...interface{}) string {
	return fs.storageService.fmt("/FileSystems/%s", fs.id) + fmt.Sprintf(format, a...)
}

func (fs *FileSystem) findFileShare(fileShareId string) *FileShare {
	for fileShareIdx, fileShare := range fs.shares {
		if fileShare.id == fileShareId {
			return &fs.shares[fileShareIdx]
		}
	}

	return nil
}

func (fs *FileSystem) createFileShare(endpoint *Endpoint, mountRoot string) *FileShare {
	var fileShareId = -1
	for _, fileShare := range fs.shares {
		id, _ := strconv.Atoi(fileShare.id)

		if fileShareId <= id {
			fileShareId = id
		}
	}

	fileShareId = fileShareId + 1

	return &FileShare{
		id:        strconv.Itoa(fileShareId),
		endpoint:  endpoint,
		mountRoot: mountRoot,
	}
}

func Initialize(ctrl NnfControllerInterface) error {

	storageService = StorageService{
		id: DefaultStorageServiceId,
	}

	s := &storageService

	conf, err := loadConfig()
	if err != nil {
		log.WithError(err).Errorf("Failed to load %s configuration")
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
			id:    strconv.Itoa(endpointIdx),
			state: sf.UNAVAILABLE_OFFLINE_RST,
		}
	}

	s.resourceIndex = strings.Count(s.fmt("/StoragePool/0"), "/") - 1

	PortEventManager.Subscribe(PortEventSubscriber{
		HandlerFunc: PortEventHandler,
		Data:        s,
	})

	return nil
}

func PortEventHandler(event PortEvent, data interface{}) {
	s := data.(*StorageService)

	if event.PortType != PORT_TYPE_USP {
		return
	}

	idx, err := api.FabricController.ConvertPortEventToRelativePortIndex(event)
	if err != nil {
		log.WithError(err).Errorf("Unable to find port index for event %+v", event)
		return
	}

	if !(idx < len(s.endpoints)) {
		log.Errorf("Port index %d exceeds endpoint count for event %+v", idx, event)
		return
	}

	controllerIndex, err := nvme.ConvertRelativePortIndexToControllerIndex(idx)
	if err != nil {
		log.WithError(err).Errorf("Could not convert port index to controller index")
		return
	}

	endpoint := &s.endpoints[idx]
	endpoint.controllerIndex = controllerIndex
	switch event.EventType {
	case PORT_EVENT_UP:
		endpoint.state = sf.ENABLED_RST
	default:
		endpoint.state = sf.UNAVAILABLE_OFFLINE_RST
	}
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
	model.StoragePools = sf.OdataV4IdRef{OdataId: s.fmt("/StorgePools")}
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

func StorageServiceIdStoragePoolPost(storageServiceId string, model *sf.StoragePoolV150StoragePool) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrBadRequest
	}

	// TODO: Check the model for valid RAID configurations

	policy := NewAllocationPolicy(s.config.AllocationConfig, model.Oem)
	if policy == nil {
		log.Errorf("Failed to allocate storage policy.")
		return ec.ErrInternalServerError
	}

	if err := policy.Initialize(uint64(model.Capacity.Data.AllocatedBytes)); err != nil {
		log.WithError(err).Errorf("Failed to initialize storage policy")
		return ec.ErrInternalServerError
	}

	if !policy.CheckCapacity() {
		log.Warnf("Storage Policy does not provide sufficient capacity to support requested bytes")
		return ec.ErrNotAcceptable
	}

	volumes, err := policy.Allocate()
	if err != nil {
		log.WithError(err).Errorf("Storage Policy allocation failed.")
		return ec.ErrInternalServerError
	}

	p := s.createStoragePool(policy, volumes)
	s.pools = append(s.pools, *p)

	model.Id = p.id
	model.OdataId = p.fmt("")
	model.AllocatedVolumes.OdataId = p.fmt("/AllocatedVolumes")

	model.CapacitySources = p.capacitySourcesGet()
	model.CapacitySourcesodataCount = int64(len(model.CapacitySources))

	return nil
}

func StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrBadRequest
	}

	model.Id = p.id
	model.AllocatedVolumes = sf.OdataV4IdRef{OdataId: p.fmt("/AllocatedVolumes")}

	model.BlockSizeBytes = 4096 // TODO
	model.Capacity = sf.CapacityV100Capacity{
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes:   p.allocatedVolume.capacityBytes,
			ProvisionedBytes: p.allocatedVolume.capacityBytes,
			ConsumedBytes:    0, // TODO
		},
	}

	model.CapacitySources = p.capacitySourcesGet()
	model.CapacitySourcesodataCount = int64(len(model.CapacitySources))

	return nil
}

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

func StorageServiceIdStoragePoolIdCapacitySourcesGet(storageServiceId, storagePoolId string, model *sf.CapacitySourceCollectionCapacitySourceCollection) error {
	_, p := findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.ErrNotFound
	}

	model.Members = p.capacitySourcesGet()
	model.MembersodataCount = int64(len(model.Members))

	return nil
}

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

	return nil
}

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

func StorageServiceIdStorageGroupPost(storageServiceId string, model *sf.StorageGroupV150StorageGroup) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	if model.MappedVolumes == nil || len(model.MappedVolumes) != 1 {
		return ec.ErrBadRequest
	}

	// TODO: Do Mapped Volume Validation
	volume := &AllocatedVolume{}

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
			return ec.ErrBadRequest
		}

		endpoints = append(endpoints, e)
	}

	// Everything validated OK - create the Storage Group

	sg := s.createStorageGroup(volume, endpoints)
	s.groups = append(s.groups, *sg)

	model.Id = sg.id
	model.OdataId = sg.fmt("")

	return nil
}

func StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	model.MappedVolumes = []sf.StorageGroupMappedVolume{
		{
			AccessCapability: sf.READ_WRITE_SGAC,
			Volume:           sf.OdataV4IdRef{OdataId: sg.volume.storagePool.fmt("/AllocatedVolume/%s", sg.volume.id)},
		},
	}

	model.ServerEndpointsodataCount = int64(len(sg.endpoints))
	model.ServerEndpoints = make([]sf.OdataV4IdRef, model.ServerEndpointsodataCount)
	for endpointIdx, endpoint := range sg.endpoints {
		model.ServerEndpoints[endpointIdx] = sf.OdataV4IdRef{OdataId: sg.fmt("/ServerEndpoints/%s", endpoint.id)}
	}

	return nil
}

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

func StorageServiceIdStorageGroupIdExposeVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150ExposeVolumes) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	controllers := make([]uint16, len(sg.endpoints))
	for endpointIdx, endpoint := range sg.endpoints {
		controllers[endpointIdx] = uint16(endpoint.controllerIndex)
	}

	sp := sg.volume.storagePool
	for _, volume := range sp.providingVolumes {
		if err := nvme.AttachControllers(volume.volume, controllers); err != nil {
			// TODO: Rollback
			return err
		}
	}

	// TODO: Change the sg resource status?

	return nil
}

func StorageServiceIdStorageGroupIdHideVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150HideVolumes) error {
	_, sg := findStorageGroup(storageServiceId, storageGroupId)
	if sg == nil {
		return ec.ErrNotFound
	}

	// TODO: Not Yet Implemented

	return nil
}

func StorageServiceIdEndpointsGet(storageServiceId string, model *sf.EndpointCollectionEndpointCollection) error {
	s := findStorageService(storageServiceId)
	if s == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(s.endpoints))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, endpoint := range s.endpoints {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints/%s", endpoint.id)}
	}

	return nil
}

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

	fs := s.createFileSystem(sp)
	s.fileSystems = append(s.fileSystems, *fs)

	model.Id = fs.id
	model.OdataId = fs.fmt("")

	return nil
}

func StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	s, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	model.CapacityBytes = fs.storagePool.allocatedVolume.capacityBytes
	model.StoragePool = sf.OdataV4IdRef{OdataId: s.fmt("/StoragePools/%s", fs.storagePool.id)}
	model.ExportedShares = sf.OdataV4IdRef{OdataId: fs.fmt("/ExportedShares")}

	return nil
}

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

func StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId string, model *sf.FileShareCollectionFileShareCollection) error {
	_, fs := findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.ErrNotFound
	}

	model.MembersodataCount = int64(len(fs.shares))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, sh := range fs.shares {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: fs.fmt("/ExportedShares/%s", sh.id)}
	}
	return nil
}

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

	sh := fs.createFileShare(ep, model.FileSharePath)
	fs.shares = append(fs.shares, *sh)

	model.Id = sh.id
	model.OdataId = fs.fmt("/ExportedShares/%s", sh.id)

	return nil
}

func StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) error {
	s, fs, sh := findExportedShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.ErrNotFound
	}

	model.FileSharePath = sh.mountRoot
	model.ServerEndpoint = sf.OdataV4IdRef{OdataId: s.fmt("/Endpoints/%s", sh.endpoint.id)}
	model.Links.FileSystem = sf.OdataV4IdRef{OdataId: fs.fmt("")}

	//model.Status // TODO

	return nil
}

func StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId string) error {
	_, fs, sh := findExportedShare(storageServiceId, fileSystemId, exportedShareId)
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
