
package server

import (
	"github.com/google/uuid"
)

type RemoteServerController struct {

}

func (*RemoteServerController) NewServerStoragePool(pid uuid.UUID) *ServerStoragePool {
	return nil
}

func (*RemoteServerController) GetStatus(pool *ServerStoragePool) ServerStoragePoolStatus {
	return ServerStoragePoolError
}

func (*RemoteServerController) CreateFileSystem(pool *ServerStoragePool, fs FileSystemApi, mountpoint string) error {
	return nil
}