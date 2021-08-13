package server

import (
	"testing"

	"github.com/google/uuid"
)

func prepare(t *testing.T, pid uuid.UUID) (*Storage, FileSystemControllerApi) {
	fsConfig, _ := loadConfig()
	fsCtrl := NewFileSystemController(fsConfig)

	provider := DefaultServerControllerProvider{}
	ctrl := provider.NewServerController(ServerControllerOptions{Local: true})
	pool := ctrl.NewStorage(pid)

	if pool == nil {
		t.Fatalf("Could not allocate storage pool %s", pid)
	}

	pool.GetStatus()

	return pool, fsCtrl
}

func _TestFileSystemZfs(t *testing.T) {
	// Warning, this assume the existence of some NVMe devices

	pool, ctrl := prepare(t, uuid.MustParse("00000000-0000-0000-0000-000000000000"))

	oem := FileSystemOem{Name: "test", Type: "ZFS"}
	fs := ctrl.NewFileSystem(oem)
	if fs == nil {
		t.Fatalf("Could not allocate ZFS file system")
	}

	if err := pool.CreateFileSystem(fs, nil); err != nil {
		t.Errorf("Create ZFS file system failed %s", err)
	}
}

func _TestFileSystemLvm(t *testing.T) {
	// Warning, this assumes the existence of some NVMe devies
	// labeled with a UUID. To retrieve the UUID, run the following
	//    nvme get-feature --namespace-id=1 --feature=0x7F /dev/nvme0n1
	//  The UUID will be located at byte offset 0x12

	pool, ctrl := prepare(t, uuid.MustParse("5d340875-a5c2-4e86-9406-37751954c022"))

	oem := FileSystemOem{Name: "test", Type: "LVM"}
	fs := ctrl.NewFileSystem(oem)
	if fs == nil {
		t.Fatal("Could not create LVM file system")
	}

	if err := pool.CreateFileSystem(fs, nil); err != nil {
		t.Fatalf("Create LVM file system failed %s", err)
	} else {
		t.Log("Create LVM file system success")
	}

	if err := pool.DeleteFileSystem(fs); err != nil {
		t.Fatalf("Delete LVM file system failed %s", err)
	} else {
		t.Log("Delete LVM file system success")
	}

}
