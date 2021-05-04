package nnf

import (
	"encoding/json"
	"fmt"
	"net/http"

	//nnf "stash.us.cray.com/rabsw/nnf-ec/internal/manager-nnf"

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

func (s *storageService) ProvisionStorage()

func (s *storageService) get(path string, model interface{}) error {
	
	rsp, err := s.client.Get(fmt.Sprintf("http://%s:%s%s", s.address, s.port, path))
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
