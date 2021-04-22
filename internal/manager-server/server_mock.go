package server

import "github.com/google/uuid"

type MockServerControllerProvider struct{}

func (MockServerControllerProvider) NewServerController(ServerControllerOptions) ServerControllerApi {
	return NewMockServerController()
}

func NewMockServerController() ServerControllerApi {
	return &MockServerController{}
}

type MockServerController struct {
}

func (c *MockServerController) NewServerStoragePool(pid uuid.UUID) *ServerStoragePool {
	return &ServerStoragePool{
		Id:   pid,
		ctrl: c,
	}
}

func (*MockServerController) GetStatus(pool *ServerStoragePool) ServerStoragePoolStatus {
	return ServerStoragePoolReady
}

func (*MockServerController) CreateFileSystem(pool *ServerStoragePool, fs FileSystemApi, mountpoint string) error {
	return nil
}
