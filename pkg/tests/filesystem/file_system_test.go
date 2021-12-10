package filesystem

import (
	"testing"

	ec "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg"
	nnf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-nnf"
	server "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-server"

	openapi "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/rfsf/pkg/common"
	sf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/rfsf/pkg/models"
)

func TestFileSystem(t *testing.T) {

	c := ec.NewController(ec.NewMockOptions())
	defer c.Close()

	if err := c.Init(nil); err != nil {
		t.Fatalf("Failed to start nnf controller")
	}

	testFs := &testFileSystem{t: t}
	server.FileSystemRegistry.RegisterFileSystem(testFs)
	// TODO: defer server.FileSystemRegistry.UnregisterFileSystem(testFs)

	ss := nnf.NewDefaultStorageService()

	sp := &sf.StoragePoolV150StoragePool{
		CapacityBytes: 1024 * 1024,
		Oem: openapi.MarshalOem(nnf.AllocationPolicyOem{
			Policy:     nnf.SpareAllocationPolicyType,
			Compliance: nnf.RelaxedAllocationComplianceType,
		}),
	}

	if err := ss.StorageServiceIdStoragePoolsPost(ss.Id(), sp); err != nil {
		t.Fatalf("Failed to create storage pool Error: %+v", err)
	}

	rabbitEndpointId := "0"
	ep := &sf.EndpointV150Endpoint{}
	if err := ss.StorageServiceIdEndpointIdGet(ss.Id(), rabbitEndpointId, ep); err != nil {
		t.Fatalf("Failed to get endpoint ID: %s Error: %+v", rabbitEndpointId, err)
	}

	sg := &sf.StorageGroupV150StorageGroup{
		Links: sf.StorageGroupV150Links{
			StoragePool:    sf.OdataV4IdRef{OdataId: sp.OdataId},
			ServerEndpoint: sf.OdataV4IdRef{OdataId: ep.OdataId},
		},
	}

	if err := ss.StorageServiceIdStorageGroupPost(ss.Id(), sg); err != nil {
		t.Fatalf("Failed to create storage group Pool ID: %s Error: %+v", sp.Id, err)
	}

	fs := &sf.FileSystemV122FileSystem{
		Links: sf.FileSystemV122Links{
			StoragePool: sf.OdataV4IdRef{OdataId: sp.OdataId},
		},
		Oem: openapi.MarshalOem(server.FileSystemOem{
			Type: testFs.Type(),
			Name: testFs.Name(),
		}),
	}

	if err := ss.StorageServiceIdFileSystemsPost(ss.Id(), fs); err != nil {
		t.Fatalf("Failed to create file system Pool ID: %s Error: %+v", sp.Id, err)
	}

	sh := &sf.FileShareV120FileShare{
		FileSharePath: "/mnt/test",
		Links: sf.FileShareV120Links{
			FileSystem: sf.OdataV4IdRef{OdataId: fs.OdataId},
			Endpoint:   sf.OdataV4IdRef{OdataId: ep.OdataId},
		},
	}

	if err := ss.StorageServiceIdFileSystemIdExportedSharesPost(ss.Id(), fs.Id, sh); err != nil {
		t.Fatalf("Failed to create file share Pool ID: %s Error: %+v", sp.Id, err)
	}

	t.Logf("Created Storage Pool %s and all subsequent resources", sp.Id)

	if testFs.created != true {
		t.Fatalf("Test File System never created")
	}

	t.Logf("Deleting Storage Pool %s", sp.Id)
	if err := ss.StorageServiceIdStoragePoolIdDelete(ss.Id(), sp.Id); err != nil {
		t.Fatalf("Failed to delete storage pool %s Error: %+v", sp.Id, err)
	}
}

type testFileSystem struct {
	master *testFileSystem
	t      *testing.T

	created    bool
	devices    []string
	mountpoint string
}

func (t *testFileSystem) New(oem server.FileSystemOem) server.FileSystemApi {
	return &testFileSystem{master: t, t: t.t, created: false, devices: nil, mountpoint: ""}
}

func (*testFileSystem) IsType(oem server.FileSystemOem) bool { return oem.Type == "test" }
func (*testFileSystem) IsMockable() bool                     { return true }

func (*testFileSystem) Type() string { return "test" }
func (*testFileSystem) Name() string { return "test" }

func (fs *testFileSystem) Create(devices []string, options server.FileSystemOptions) error {
	fs.t.Logf("Test File System: Create File System: Devices: %+v Options: %+v", devices, options)

	if devices == nil {
		fs.t.Fatalf("Test File System: Expected non-nil device list")
	}

	if fs.devices != nil {
		fs.t.Errorf("Test File System: Create of file system already has device list: Devices: %+v", fs.devices)
	}

	fs.created = true
	fs.devices = devices

	fs.master.created = true
	return nil
}

func (fs *testFileSystem) Delete() error {
	fs.t.Logf("Test File System: Delete File System")

	if fs.created == false {
		fs.t.Errorf("Test File System: Delete has not detected create")
	}

	if fs.devices == nil {
		fs.t.Errorf("Test File System: Delete has no devices present. Could be a duplicate delete.")
	}

	fs.devices = nil
	return nil
}

func (fs *testFileSystem) Mount(mountpoint string) error {
	fs.t.Logf("Test File System: Mount %s", mountpoint)

	if mountpoint == "" {
		fs.t.Fatal("Mount requires non empty mountpoint")
	}
	if fs.mountpoint != "" {
		fs.t.Errorf("Teste File System: Mountpoint already present: Mountpoint: %s", fs.mountpoint)
	}

	fs.mountpoint = mountpoint
	return nil
}

func (fs *testFileSystem) Unmount() error {
	fs.t.Logf("Test File System: Unmount")

	if fs.mountpoint == "" {
		fs.t.Errorf("Test File System: Unmount has no mountpoint")
	}

	fs.mountpoint = ""
	return nil
}
