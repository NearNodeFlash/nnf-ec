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

package recovery_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nnfec "github.com/NearNodeFlash/nnf-ec/pkg"
	ec "github.com/NearNodeFlash/nnf-ec/pkg/ec"

	nnf "github.com/NearNodeFlash/nnf-ec/pkg/manager-nnf"
	nnfserver "github.com/NearNodeFlash/nnf-ec/pkg/manager-server"

	openapi "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/common"
	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
)

/*
Test of Reboot / Crash Recovery. This test will, one by one, test the creation of objects based on the object
chain we have laid out for Rabbit.

	Storage Pool -> Storage Group -> File System -> File Share

We do an additional read of the Server Endpoints to acquire the rabbit endpoint id; so the final chain
looks as follows

	Storage Pool -> Server Endpoint -> Storage Group -> File System -> File Share

We test the chain one by one, so the first test iteration will test only Storage Pool recovery; the next
iteration will test Storage Pool and Storage Group recovery; etc etc. This is useful in testing the bare
minimum of object recovery before building on top of the previous work.

Each object must support Create, Verify, and Delete APIs. Objects should also return the Id() of the resource
that was created. This way objects in the chain can access the ID of previous creates. This means the chain
order must be preserved as to make access predictable. For example, creating a file system requires a
Storage Pool ID, so the file system object API assumes chain index 0 is a Storage Pool
*/
var _ = Describe("Reboot Recovery Testing", func() {

	objectApis := [...]testRecoveryApi{
		newTestStoragePoolObjectApi(),
		newTestServerEndpointObjectApi(), // Not actually created, used only to read a server endpoint
		newTestStorageGroupObjectApi(),
		newTestFileSystemObjectApi(),
		newTestFileShareObjectApi(),
	}

	// Test object creation up until objIdx + 1
	for objIdx := range objectApis {

		objIdx := objIdx
		obj := objectApis[objIdx]

		var c *ec.Controller
		var ss nnf.StorageServiceApi

		Describe(fmt.Sprintf("Create and Recover Object %s", obj.Name()), func() {

			// Before each test we start the NNF Element Controller. This will initialize
			// everything, including the persistent database. If any data is in the DB, it
			// will be recovered
			BeforeEach(func() {
				c = nnfec.NewController(nnfec.NewMockOptions(true))
				Expect(c.Init(ec.NewDefaultOptions())).NotTo(HaveOccurred())

				ss = nnf.NewDefaultStorageService()
			})

			// After each test we close the NNF Element Controller, thereby safely closing
			// the persistent database.
			AfterEach(func() {
				c.Close()
			})

			// Create the chain of objects up until the stop value
			It(fmt.Sprintf("Creates the object chain at index %d (%s)", objIdx, obj.Name()), func() {
				for _, objApi := range objectApis[:objIdx+1] {
					Expect(objApi.CreateObject(ss, objectApis[:]...)).Should(Succeed())
					Expect(objApi.VerifyObject(ss)).Should(Succeed())
				}
			})

			It(fmt.Sprintf("Recovers the object chain at index %d (%s)", objIdx, obj.Name()), func() {
				for _, objApi := range objectApis[:objIdx+1] {
					Expect(objApi.VerifyObject(ss)).Should(Succeed())
				}

				// TODO: Delete the object in reverse order
			})
		})
	}
})

type testRecoveryApi interface {
	Name() string // Returns a simple name for this interface

	CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error
	VerifyObject(ss nnf.StorageServiceApi) error
	DeleteObject(ss nnf.StorageServiceApi) error

	// Return the ID or OdataId of the created object
	Id() string
	OdataId() string
}

const (
	StoragePoolObjectName    = "Storage Pool"
	ServerEndpointObjectName = "Server Endpoint"
	StorageGroupObjectName   = "Storage Group"
	FileSystemObjectName     = "File System"
	FileShareObjectName      = "File Share"
)

func newInvalidObjectError(idx int, expected string) error {
	return fmt.Errorf("invalid object in object chain. Index: %d Expected: %s", idx, expected)
}

/*
Storage Pool
*/

func newTestStoragePoolObjectApi() testRecoveryApi {
	return &testStoragePoolObject{}
}

type testStoragePoolObject struct {
	sp sf.StoragePoolV150StoragePool
}

func (o *testStoragePoolObject) Name() string { return StoragePoolObjectName }

func (o *testStoragePoolObject) Id() string      { return o.sp.Id }
func (o *testStoragePoolObject) OdataId() string { return o.sp.OdataId }

func (o *testStoragePoolObject) CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error {
	o.sp = sf.StoragePoolV150StoragePool{
		CapacityBytes: 1024 * 1024 * 1024, // 1 GiB
		Oem: openapi.MarshalOem(nnf.AllocationPolicyOem{
			Policy:     nnf.SpareAllocationPolicyType,
			Compliance: nnf.RelaxedAllocationComplianceType,
		}),
	}

	return ss.StorageServiceIdStoragePoolsPost(ss.Id(), &o.sp)
}

func (o *testStoragePoolObject) VerifyObject(ss nnf.StorageServiceApi) error {
	var sp sf.StoragePoolV150StoragePool
	if err := ss.StorageServiceIdStoragePoolIdGet(ss.Id(), o.Id(), &sp); err != nil {
		return err
	}

	// TODO: Verify the returned object matches the created object
	return nil
}

func (o *testStoragePoolObject) DeleteObject(ss nnf.StorageServiceApi) error {
	return ss.StorageServiceIdStoragePoolIdDelete(ss.Id(), o.sp.Id)
}

/*
Server Endpoint - this is just used to read the server endpoints; no creation / verification / deletion is actually done
*/

func newTestServerEndpointObjectApi() testRecoveryApi {
	return &testServerEndpointObject{}
}

type testServerEndpointObject struct {
	ep sf.EndpointV150Endpoint
}

func (o *testServerEndpointObject) Name() string { return ServerEndpointObjectName }

func (o *testServerEndpointObject) Id() string      { return o.ep.Id }
func (o *testServerEndpointObject) OdataId() string { return o.ep.OdataId }

func (o *testServerEndpointObject) CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error {
	return ss.StorageServiceIdEndpointIdGet(ss.Id(), "0", &o.ep)
}

func (o *testServerEndpointObject) VerifyObject(nnf.StorageServiceApi) error { return nil }
func (o *testServerEndpointObject) DeleteObject(nnf.StorageServiceApi) error { return nil }

/*
Storage Group
*/

func newTestStorageGroupObjectApi() testRecoveryApi {
	return &testStorageGroupObject{}
}

type testStorageGroupObject struct {
	sg sf.StorageGroupV150StorageGroup
}

func (o *testStorageGroupObject) Name() string { return StorageGroupObjectName }

func (o *testStorageGroupObject) Id() string      { return o.sg.Id }
func (o *testStorageGroupObject) OdataId() string { return o.sg.OdataId }

func (o *testStorageGroupObject) CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error {
	if objs[0].Name() != StoragePoolObjectName {
		return newInvalidObjectError(0, StoragePoolObjectName)
	}
	if objs[1].Name() != ServerEndpointObjectName {
		return newInvalidObjectError(1, ServerEndpointObjectName)
	}

	o.sg = sf.StorageGroupV150StorageGroup{
		Links: sf.StorageGroupV150Links{
			StoragePool:    sf.OdataV4IdRef{OdataId: objs[0].OdataId()},
			ServerEndpoint: sf.OdataV4IdRef{OdataId: objs[1].OdataId()},
		},
	}

	return ss.StorageServiceIdStorageGroupPost(ss.Id(), &o.sg)
}

func (o *testStorageGroupObject) VerifyObject(ss nnf.StorageServiceApi) error {
	var sg sf.StorageGroupV150StorageGroup
	if err := ss.StorageServiceIdStorageGroupIdGet(ss.Id(), o.sg.Id, &sg); err != nil {
		return err
	}

	return nil
}

func (o *testStorageGroupObject) DeleteObject(ss nnf.StorageServiceApi) error {
	return ss.StorageServiceIdStorageGroupIdDelete(ss.Id(), o.sg.Id)
}

/*
File System
*/

func newTestFileSystemObjectApi() testRecoveryApi {
	return &testFileSystemObject{}
}

type testFileSystemObject struct {
	fs sf.FileSystemV122FileSystem
}

func (o *testFileSystemObject) Name() string { return FileSystemObjectName }

func (o *testFileSystemObject) Id() string      { return o.fs.Id }
func (o *testFileSystemObject) OdataId() string { return o.fs.OdataId }

func (o *testFileSystemObject) CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error {
	o.fs = sf.FileSystemV122FileSystem{
		Links: sf.FileSystemV122Links{
			StoragePool: sf.OdataV4IdRef{OdataId: objs[0].OdataId()},
		},
		Oem: openapi.MarshalOem(
			nnfserver.FileSystemOem{
				Type: "lvm",
				Name: "lvm",
			},
		),
	}

	return ss.StorageServiceIdFileSystemsPost(ss.Id(), &o.fs)
}

func (o *testFileSystemObject) VerifyObject(ss nnf.StorageServiceApi) error {
	return nil
}

func (o *testFileSystemObject) DeleteObject(ss nnf.StorageServiceApi) error {
	return ss.StorageServiceIdFileSystemIdDelete(ss.Id(), o.fs.Id)
}

/*
File Share
*/

func newTestFileShareObjectApi() testRecoveryApi {
	return &testFileShareObject{}
}

type testFileShareObject struct {
	sh   sf.FileShareV120FileShare
	fsId string
}

func (o *testFileShareObject) Name() string { return FileShareObjectName }

func (o *testFileShareObject) Id() string      { return o.sh.Id }
func (o *testFileShareObject) OdataId() string { return o.sh.OdataId }

func (o *testFileShareObject) CreateObject(ss nnf.StorageServiceApi, objs ...testRecoveryApi) error {
	if objs[4].Name() != FileShareObjectName {
		return newInvalidObjectError(4, FileShareObjectName)
	}

	o.sh = sf.FileShareV120FileShare{
		FileSharePath: "/mnt/rabbit",
		Links: sf.FileShareV120Links{
			Endpoint:   sf.OdataV4IdRef{OdataId: objs[1].OdataId()},
			FileSystem: sf.OdataV4IdRef{OdataId: objs[3].OdataId()},
		},
	}

	o.fsId = objs[3].Id()

	return ss.StorageServiceIdFileSystemIdExportedSharesPost(ss.Id(), o.fsId, &o.sh)
}

func (o *testFileShareObject) VerifyObject(ss nnf.StorageServiceApi) error { return nil }

func (o *testFileShareObject) DeleteObject(ss nnf.StorageServiceApi) error {
	return ss.StorageServiceIdFileSystemIdExportedShareIdDelete(ss.Id(), o.fsId, o.sh.Id)
}
