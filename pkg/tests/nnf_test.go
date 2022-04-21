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

package benchmarks

import (
	"strings"
	"testing"

	ec "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg"
	nnf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-nnf"
	server "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-server"

	openapi "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/rfsf/pkg/common"
	sf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/rfsf/pkg/models"
)

func TestStoragePools(t *testing.T) {
	c := ec.NewController(ec.NewMockOptions(false))
	defer c.Close()

	if err := c.Init(nil); err != nil {
		t.Fatalf("Failed to start nnf controller")
	}

	ss := nnf.NewDefaultStorageService()

	cs := &sf.CapacityCapacitySource{}
	if err := ss.StorageServiceIdCapacitySourceGet(ss.Id(), cs); err != nil {
		t.Errorf("Failed to retrieve capacity source: %v", err)
	}

	// Retrieve the list of endpoints in the same way nnf-sos queries the list of servers
	// for status. This can bringout bugs where the server is down.
	{
		epc := &sf.EndpointCollectionEndpointCollection{}
		if err := ss.StorageServiceIdEndpointsGet(ss.Id(), epc); err != nil {
			t.Errorf("Failed to retrieve endpoints: %v", err)
		}

		for _, ref := range epc.Members {
			epid := ref.OdataId[strings.LastIndex(ref.OdataId, "/")+1:]

			ep := &sf.EndpointV150Endpoint{}
			if err := ss.StorageServiceIdEndpointIdGet(ss.Id(), epid, ep); err != nil {
				t.Errorf("Failed to retrieve endpoint id %s: %v", epid, err)
			}
		}
	}

	pools := make([]*sf.StoragePoolV150StoragePool, 0)

	for j := 0; j < 32; j++ {

		sp := &sf.StoragePoolV150StoragePool{
			CapacityBytes: 1024 * 1024,
			Oem: openapi.MarshalOem(nnf.AllocationPolicyOem{
				Policy:     nnf.SpareAllocationPolicyType,
				Compliance: nnf.RelaxedAllocationComplianceType,
			}),
		}

		if err := ss.StorageServiceIdStoragePoolsPost(ss.Id(), sp); err != nil {
			t.Fatalf("Failed to create storage pool %d Error: %+v", j, err)
		}

		pools = append(pools, sp)

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
				Type: "zfs",
				Name: "zfs",
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
	}

	t.Logf("Created %d Storage Pools. Starting Delete...", len(pools))

	for _, pool := range pools {
		if err := ss.StorageServiceIdStoragePoolIdDelete(ss.Id(), pool.Id); err != nil {
			t.Fatalf("Failed to delete storage pool ID %s Error: %v", pool.Id, err)
		}
	}
}
