package nnf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	nnf "stash.us.cray.com/rabsw/nnf-ec/internal/manager-nnf"
	server "stash.us.cray.com/rabsw/nnf-ec/internal/manager-server"

	openapi "stash.us.cray.com/rabsw/rfsf-openapi/pkg/common"
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

func Connect(address, port string) *storageService {
	return &storageService{
		address: address,
		port:    port,
		client:  http.Client{},
	}
}

type storageService struct {
	address string
	port    string
	client  http.Client
}

func (s *storageService) Get() (*sf.StorageServiceV150StorageService, error) {
	model := new(sf.StorageServiceV150StorageService)
	err := s.get("/redfish/v1/StorageServices/NNF", model)

	return model, err
}

func (s *storageService) Capacity() (*sf.CapacityCapacitySource, error) {
	model := new(sf.CapacityCapacitySource)
	err := s.get("/redfish/v1/StorageServices/NNF/CapacitySource", model)

	return model, err
}

func (s *storageService) AllocateStorage(capacityBytes int64) (*sf.StoragePoolV150StoragePool, error) {
	model := new(sf.StoragePoolV150StoragePool)

	model.CapacityBytes = capacityBytes
	model.Oem = openapi.MarshalOem(nnf.AllocationPolicyOem{
		Policy:     nnf.SpareAllocationPolicyType,
		Compliance: nnf.RelaxedAllocationComplianceType,
	})

	err := s.post("/redfish/v1/StorageServices/NNF/StoragePools", model)

	return model, err
}

func (s *storageService) CreateStorageGroup(pool *sf.StoragePoolV150StoragePool, endpoint *sf.EndpointV150Endpoint) (*sf.StorageGroupV150StorageGroup, error) {
	model := new(sf.StorageGroupV150StorageGroup)

	model.Links.StoragePool.OdataId = pool.OdataId
	model.Links.ServerEndpoint.OdataId = endpoint.OdataId

	err := s.post("/redfish/v1/StorageServices/NNF/StorageGroups", model)

	return model, err
}

func (s *storageService) CreateFileSystem(pool *sf.StoragePoolV150StoragePool, fileSystem string) (*sf.FileSystemV122FileSystem, error) {
	model := new(sf.FileSystemV122FileSystem)

	model.Links.StoragePool.OdataId = pool.OdataId
	model.Oem = openapi.MarshalOem(server.FileSystemOem{
		Name: fileSystem,
	})

	err := s.post("/redfish/v1/StorageServices/NNF/FileSystems", model)

	return model, err
}

func (s *storageService) CreateFileShare(fileSystem *sf.FileSystemV122FileSystem, endpoint *sf.EndpointV150Endpoint, fileSharePath string) (*sf.FileShareV120FileShare, error) {
	model := new(sf.FileShareV120FileShare)

	model.FileSharePath = fileSharePath
	model.Links.FileSystem.OdataId = fileSystem.OdataId
	model.Links.Endpoint.OdataId = endpoint.OdataId

	err := s.post(fileSystem.OdataId+"/ExportedFileShares", model)

	return model, err
}

func (s *storageService) get(path string, model interface{}) error {
	return s.do(http.MethodGet, path, model)
}

func (s *storageService) post(path string, model interface{}) error {
	return s.do(http.MethodPost, path, model)
}

func (s *storageService) do(method string, path string, model interface{}) error {
	url := fmt.Sprintf("http://%s:%s%s", s.address, s.port, path)

	body := []byte{}
	if method == http.MethodPost || method == http.MethodPatch {
		body, _ = json.Marshal(model)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	rsp, err := s.client.Do(req)
	if rsp != nil {
		defer rsp.Body.Close()
	}

	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("Get request failed. Path: %s Status: %d (%s)", path, rsp.StatusCode, rsp.Status)
	}

	if err := json.NewDecoder(rsp.Body).Decode(model); err != nil {
		return err
	}

	return nil
}
