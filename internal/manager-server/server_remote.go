package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

const (
	RemoteStorageServiceId   = "NNFServer"
	RemoteStorageServicePort = 60050
)

type RemoteServerController struct {
	Address string
}

func (c *RemoteServerController) NewServerStoragePool(pid uuid.UUID) *ServerStoragePool {

	pool := sf.StoragePoolV150StoragePool{
		Id: pid.String(),
	}

	req, _ := json.Marshal(pool)

	rsp, err := http.Post(
		c.Url("/StoragePools"),
		"application/json",
		bytes.NewBuffer(req),
	)

	if err != nil {
		log.WithError(err).Errorf("New Server Storage: Http Error")
		return nil
	}

	defer rsp.Body.Close()

	if err := json.NewDecoder(rsp.Body).Decode(&pool); err != nil {
		log.WithError(err).Errorf("New Server Storage: Failed to decode JSON response")
		return nil
	}

	return &ServerStoragePool{
		Id:   pid,
		ctrl: c,
	}
}

func (c *RemoteServerController) GetStatus(p *ServerStoragePool) ServerStoragePoolStatus {

	rsp, err := http.Get(
		c.Url(fmt.Sprintf("/StoragePools/%s", p.Id.String())),
	)
	if err != nil {
		log.WithError(err).Errorf("Get Status: Http Error")
		return ServerStoragePoolError
	}

	defer rsp.Body.Close()

	pool := sf.StoragePoolV150StoragePool{}
	if err := json.NewDecoder(rsp.Body).Decode(&pool); err != nil {
		log.WithError(err).Errorf("Get Status: Failed to decode JSON response")
		return ServerStoragePoolError
	}

	switch pool.Status.State {
	case sf.STARTING_RST:
		return ServerStoragePoolStarting
	case sf.ENABLED_RST:
		return ServerStoragePoolReady
	default:
		return ServerStoragePoolError
	}
}

func (c *RemoteServerController) CreateFileSystem(p *ServerStoragePool, f FileSystemApi, mp string) error {

	fileSystem := sf.FileSystemV122FileSystem{
		StoragePool: sf.OdataV4IdRef{OdataId: fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools/%s", RemoteStorageServiceId, p.Id.String())},
		Oem:         map[string]interface{}{"FileSystem": FileSystemOem{Name: f.Name()}},
	}

	req, _ := json.Marshal(fileSystem)

	rsp, err := http.Post(
		c.Url("/FileSystems"),
		"application/json",
		bytes.NewBuffer(req),
	)

	if err != nil {
		log.WithError(err).Errorf("Create File System: Http Error")
		return err
	}

	defer rsp.Body.Close()

	if err := json.NewDecoder(rsp.Body).Decode(&fileSystem); err != nil {
		log.WithError(err).Errorf("Create File System: Failed to decode JSON response")
		return err
	}

	return c.createMountPoint(&fileSystem, mp)
}

func (c *RemoteServerController) createMountPoint(fs *sf.FileSystemV122FileSystem, mp string) error {

	fileShare := sf.FileShareV120FileShare{
		FileSharePath: mp,
	}

	req, _ := json.Marshal(fileShare)

	rsp, err := http.Post(
		c.Url(fmt.Sprintf("/FileSystems/%s", fs.Id)),
		"application/json",
		bytes.NewBuffer(req),
	)

	if err != nil {
		log.WithError(err).Errorf("Create Mount Point: Http Error")
		return err
	}

	defer rsp.Body.Close()

	if err := json.NewDecoder(rsp.Body).Decode(&fileShare); err != nil {
		log.WithError(err).Errorf("Create Mount Point: Failed to decode JSON response")
		return err
	}

	return nil
}

func (r *RemoteServerController) Url(path string) string {
	return fmt.Sprintf("http://%s:%d/redfish/v1/StorageServices/%s%s", r.Address, RemoteStorageServicePort, RemoteStorageServiceId, path)
}
