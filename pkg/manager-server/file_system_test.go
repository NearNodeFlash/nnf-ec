/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	pool := ctrl.NewStorage(pid, nil)

	if pool == nil {
		t.Fatalf("Could not allocate storage pool %s", pid)
	}

	pool.GetStatus()

	return pool, fsCtrl
}

func _TestFileSystemZfs(t *testing.T) {
	// Warning, this assume the existence of some NVMe devices

	pool, ctrl := prepare(t, uuid.MustParse("00000000-0000-0000-0000-000000000000"))

	oem := &FileSystemOem{Name: "test", Type: "ZFS"}
	fs, err := ctrl.NewFileSystem(oem)
	if err != nil {
		t.Fatal(err)
	}
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

	oem := &FileSystemOem{Name: "test", Type: "LVM"}
	fs, err := ctrl.NewFileSystem(oem)
	if err != nil {
		t.Fatal(err)
	}
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
