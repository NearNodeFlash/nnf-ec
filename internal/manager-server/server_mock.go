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

func (c *MockServerController) Connected() bool { return true }

func (c *MockServerController) NewStorage(pid uuid.UUID) *Storage {
	return &Storage{
		Id:   pid,
		ctrl: c,
	}
}

func (c *MockServerController) Delete(s *Storage) error {
	return nil
}

func (*MockServerController) GetStatus(s *Storage) StorageStatus {
	return StorageStatus_Ready
}

func (*MockServerController) CreateFileSystem(s *Storage, fs FileSystemApi, mountpoint string) error {
	return nil
}

func (*MockServerController) DeleteFileSystem(s *Storage) error {
	return nil
}