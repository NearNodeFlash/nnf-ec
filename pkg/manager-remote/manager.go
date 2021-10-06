package remote

import (
	"fmt"
	"os"

	"github.com/google/uuid"

	openapi "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/common"
	sf "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/models"

	ec "stash.us.cray.com/rabsw/nnf-ec/pkg/ec"
	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg/manager-nnf"
	server "stash.us.cray.com/rabsw/nnf-ec/pkg/manager-server"
)

type ServerStorageService struct {
	id string

	pools       []StoragePool
	fileSystems []FileSystem

	ctrl server.LocalServerController
}

type StoragePool struct {
	id string

	uid uuid.UUID

	volumes    Volume
	fileSystem *FileSystem

	serverStorageService *ServerStorageService
	serverStorage        *server.Storage
}

type Volume struct {
}

type FileSystem struct {
	id string

	api       server.FileSystemApi
	fileShare *FileShare // only one

	pool                 *StoragePool
	serverStorageService *ServerStorageService
}

type FileShare struct {
	id string

	fileSystem *FileSystem
}

func NewDefaultServerStorageService(opts *Options) nnf.StorageServiceApi {
	return &ServerStorageService{id: server.RemoteStorageServiceId}
}

func (s *ServerStorageService) OdataId() string {
	return fmt.Sprintf("/redfish/v1/StorageServices/%s", s.id)
}
func (s *ServerStorageService) OdataIdRef(ref string) sf.OdataV4IdRef {
	return sf.OdataV4IdRef{OdataId: fmt.Sprintf("%s%s", s.OdataId(), ref)}
}

func (s *ServerStorageService) isStorageService(id string) bool { return s.id == id }

func (s *ServerStorageService) findStoragePool(storageServiceId, storagePoolId string) *StoragePool {
	if !s.isStorageService(storageServiceId) {
		return nil
	}

	for poolIdx, pool := range s.pools {
		if pool.id == storagePoolId {
			return &s.pools[poolIdx]
		}
	}

	return nil
}

func (s *ServerStorageService) findFileSystem(storageServiceId, fileSystemId string) *FileSystem {
	if !s.isStorageService(storageServiceId) {
		return nil
	}

	for fsIdx, fs := range s.fileSystems {
		if fs.id == fileSystemId {
			return &s.fileSystems[fsIdx]
		}
	}

	return nil
}

func (s *ServerStorageService) findFileShare(storageServiceId, fileSystemId, fileShareId string) *FileShare {
	fs := s.findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return nil
	}

	if fs.fileShare.id == fileShareId {
		return fs.fileShare
	}

	return nil
}

func (p *StoragePool) OdataId() string {
	return p.serverStorageService.OdataId() + fmt.Sprintf("/StoragePools/%s", p.id)
}

func (p *StoragePool) OdataIdRef(ref string) sf.OdataV4IdRef {
	return sf.OdataV4IdRef{OdataId: fmt.Sprintf("%s%s", p.OdataId(), ref)}
}

func (fs *FileSystem) OdataId() string {
	return fs.serverStorageService.OdataId() + fmt.Sprintf("/FileSystems/%s", fs.id)
}

func (fs *FileSystem) OdataIdRef(ref string) sf.OdataV4IdRef {
	return sf.OdataV4IdRef{OdataId: fmt.Sprintf("%s%s", fs.OdataId(), ref)}
}

func (sh *FileShare) OdataId() string {
	return sh.fileSystem.OdataId() + fmt.Sprintf("/ExportedFileShares/%s", sh.id)
}

func (s *ServerStorageService) Id() string {
	return s.id
}

func (s *ServerStorageService) Initialize(nnf.NnfControllerInterface) error {

	// Ensure the server is initialized
	if err := server.Initialize(); err != nil {
		return err
	}

	// Perform an initial discovery of existing storage pools
	if err := s.ctrl.Discover(nil, s.NewStorage); err != nil {
		return err
	}

	// Start our monitor for receiving system events. This is what triggers updates
	// to the storage service's inner workings.

	m := UDevMonitor{}

	if err := m.Open(); err != nil {
		return err
	}

	go func() {
		events := m.Run()

		for {
			select {
			case <-m.exit:
				fmt.Println("Exiting...")
				os.Exit(0)
			case event := <-events:
				if !event.IsNvmeEvent() {
					continue
				}

				// Trigger discovery of NVMe devices
				if err := s.ctrl.Discover(nil, s.NewStorage); err != nil {
					fmt.Printf("Discover Storage Error %s\n", err)
				}
			}

		}
	}()

	return nil
}

func (s *ServerStorageService) Close() error {
	return nil
}

func (s *ServerStorageService) NewStorage(storage *server.Storage) {

	pool := StoragePool{
		id:                   storage.Id.String(),
		uid:                  storage.Id,
		serverStorage:        storage,
		serverStorageService: s,
	}

	s.pools = append(s.pools, pool)
}

func (s *ServerStorageService) StorageServicesGet(model *sf.StorageServiceCollectionStorageServiceCollection) error {
	return ec.NewErrNotAcceptable()
}

func (s *ServerStorageService) StorageServiceIdGet(storageServiceId string, model *sf.StorageServiceV150StorageService) error {
	if !s.isStorageService(storageServiceId) {
		return ec.NewErrNotFound()
	}

	model.Id = s.id
	model.OdataId = s.OdataId()
	model.StoragePools = s.OdataIdRef("/StoragePools")
	model.FileSystems = s.OdataIdRef("/FileSystems")

	return nil
}

func (s *ServerStorageService) StorageServiceIdCapacitySourceGet(storageServiceId string, model *sf.CapacityCapacitySource) error {
	return ec.NewErrNotAcceptable()
}

func (s *ServerStorageService) StorageServiceIdStoragePoolsGet(storageServiceId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	if !s.isStorageService(storageServiceId) {
		return ec.NewErrNotFound()
	}

	model.MembersodataCount = int64(len(s.pools))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for poolIdx, pool := range s.pools {
		model.Members[poolIdx] = sf.OdataV4IdRef{OdataId: pool.OdataId()}
	}

	return nil
}

func (s *ServerStorageService) StorageServiceIdStoragePoolsPost(storageServiceId string, model *sf.StoragePoolV150StoragePool) error {
	pid, err := uuid.Parse(model.Id)
	if err != nil {
		return ec.NewErrBadRequest()
	}

	for _, pool := range s.pools {
		if pool.id == model.Id {
			return nil
		}
	}

	s.NewStorage(s.ctrl.NewStorage(pid))

	return nil
}

func (s *ServerStorageService) StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	p := s.findStoragePool(storageServiceId, storagePoolId)
	if p == nil {
		return ec.NewErrNotFound()
	}

	model.Id = p.id
	model.OdataId = p.OdataId()
	//model.AllocatedVolumes = p.OdataIdRef("/AllocatedVolumes")

	// TODO: Capacity

	model.Identifier = sf.ResourceIdentifier{
		DurableName:       p.uid.String(),
		DurableNameFormat: sf.UUID_RV1100DNF,
	}

	if p.fileSystem != nil {
		model.Links.FileSystem = p.fileSystem.OdataIdRef("")
	}

	model.Status.State =
		s.ctrl.GetStatus(p.serverStorage).State()

	return nil
}

func (s *ServerStorageService) StorageServiceIdStoragePoolIdDelete(storageServiceId, storagePoolId string) error {
	sp := s.findStoragePool(storageServiceId, storagePoolId)
	if sp == nil {
		return ec.NewErrNotFound()
	}

	if sp.fileSystem != nil {
		return ec.NewErrNotAcceptable()
	}

	for poolIdx, pool := range s.pools {
		if pool.id == sp.id {
			s.pools = append(s.pools[:poolIdx], s.pools[:poolIdx+1]...)
			break
		}
	}

	return nil
}

func (*ServerStorageService) StorageServiceIdStoragePoolIdCapacitySourcesGet(storageServiceId, storagePoolId string, model *sf.CapacitySourceCollectionCapacitySourceCollection) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdStoragePoolIdCapacitySourceIdGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.CapacityCapacitySource) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdStoragePoolIdCapacitySourceIdProvidingVolumesGet(storageServiceId, storagePoolId, capacitySourceId string, model *sf.VolumeCollectionVolumeCollection) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdStoragePoolIdAlloctedVolumesGet(storageServiceId, storagePoolId string, model *sf.VolumeCollectionVolumeCollection) error {
	return nil
}

func (*ServerStorageService) StorageServiceIdStoragePoolIdAllocatedVolumeIdGet(storageServiceId, storagePoolId, volumeId string, model *sf.VolumeV161Volume) error {
	return nil
}

func (*ServerStorageService) StorageServiceIdStorageGroupsGet(storageServiceId string, model *sf.StorageGroupCollectionStorageGroupCollection) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdStorageGroupPost(storageServiceId string, model *sf.StorageGroupV150StorageGroup) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	return ec.NewErrNotAcceptable()
}

func (s *ServerStorageService) StorageServiceIdStorageGroupIdDelete(storageServiceId, storageGroupId string) error {
	sp := s.findStoragePool(storageServiceId, storageGroupId)
	if sp == nil {
		return ec.NewErrNotFound()
	}

	return nil
}

func (*ServerStorageService) StorageServiceIdEndpointsGet(storageServiceId string, model *sf.EndpointCollectionEndpointCollection) error {
	return ec.NewErrNotAcceptable()
}

func (*ServerStorageService) StorageServiceIdEndpointIdGet(storageServiceId, endpointId string, model *sf.EndpointV150Endpoint) error {
	return ec.NewErrNotAcceptable()
}

func (s *ServerStorageService) StorageServiceIdFileSystemsGet(storageServiceId string, model *sf.FileSystemCollectionFileSystemCollection) error {
	if !s.isStorageService(storageServiceId) {
		return ec.NewErrNotFound()
	}

	model.MembersodataCount = int64(len(s.fileSystems))
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	for idx, fs := range s.fileSystems {
		model.Members[idx] = sf.OdataV4IdRef{OdataId: fs.OdataId()}
	}

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemsPost(storageServiceId string, model *sf.FileSystemV122FileSystem) error {
	if !s.isStorageService(storageServiceId) {
		return ec.NewErrNotFound()
	}

	p := func() *StoragePool {
		for poolIdx, pool := range s.pools {
			if pool.OdataId() == model.StoragePool.OdataId {
				return &s.pools[poolIdx]
			}
		}
		return nil
	}()

	if p == nil {
		return ec.NewErrBadRequest()
	}

	if p.fileSystem != nil {
		return ec.NewErrBadRequest()
	}

	oem := server.FileSystemOem{}
	if err := openapi.UnmarshalOem(model.Oem, &oem); err != nil {
		return ec.NewErrBadRequest()
	}

	api := server.FileSystemController.NewFileSystem(oem)
	if api == nil {
		return ec.NewErrBadRequest()
	}

	fs := FileSystem{
		id:                   p.id, // File System Id same as Pool Id
		pool:                 p,
		api:                  api,
		serverStorageService: s,
	}

	s.fileSystems = append(s.fileSystems, fs)
	p.fileSystem = &s.fileSystems[len(s.fileSystems)-1]

	return s.StorageServiceIdFileSystemIdGet(storageServiceId, p.id, model)
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	fs := s.findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound()
	}

	model.Id = fs.id
	model.OdataId = fs.OdataId()
	model.StoragePool = sf.OdataV4IdRef{OdataId: fs.pool.OdataId()}
	//model.Capacity = fs.api.Capacity()

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdDelete(storageServiceId, fileSystemId string) error {
	fs := s.findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound()
	}

	if err := fs.api.Delete(); err != nil {
		return ec.NewErrInternalServerError()
	}

	fs.pool.fileSystem = nil

	for fileSystemIdx, fileSystem := range s.fileSystems {
		if fileSystem.id == fs.id {
			s.fileSystems = append(s.fileSystems[:fileSystemIdx], s.fileSystems[fileSystemIdx+1:]...)
			break
		}
	}

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId string, model *sf.FileShareCollectionFileShareCollection) error {
	fs := s.findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound()
	}

	if fs.fileShare == nil {
		model.MembersodataCount = 0
	} else {
		model.MembersodataCount = 1
		model.Members = []sf.OdataV4IdRef{
			{OdataId: fs.fileShare.OdataId()},
		}
	}

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId string, model *sf.FileShareV120FileShare) error {
	fs := s.findFileSystem(storageServiceId, fileSystemId)
	if fs == nil {
		return ec.NewErrNotFound()
	}

	if len(model.FileSharePath) == 0 {
		return ec.NewErrBadRequest()
	}

	if fs.fileShare != nil {
		return ec.NewErrNotAcceptable()
	}

	if err := fs.pool.serverStorage.CreateFileSystem(fs.api, model.Oem); err != nil {
		return ec.NewErrInternalServerError()
	}

	fs.fileShare = &FileShare{id: fs.id, fileSystem: fs}

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) error {
	sh := s.findFileShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.NewErrNotFound()
	}

	model.Id = sh.id
	model.OdataId = sh.OdataId()

	return nil
}

func (s *ServerStorageService) StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId string) error {
	sh := s.findFileShare(storageServiceId, fileSystemId, exportedShareId)
	if sh == nil {
		return ec.NewErrNotFound()
	}

	sh.fileSystem.fileShare = nil

	return nil
}
