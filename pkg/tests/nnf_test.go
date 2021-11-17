package benchmarks

import (
	"testing"

	ec "stash.us.cray.com/rabsw/nnf-ec/pkg"
	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg/manager-nnf"
	openapi "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/common"
	sf "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/models"
)

func TestStoragePools(t *testing.T) {
	c := ec.NewController(ec.NewMockOptions())
	defer c.Close()

	if err := c.Init(nil); err != nil {
		t.Fatalf("Failed to start nnf controller")
	}

	ss := nnf.NewDefaultStorageService()
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
	}

	for _, pool := range pools {
		if err := ss.StorageServiceIdStoragePoolIdDelete(ss.Id(), pool.Id); err != nil {
			t.Fatalf("Failed to delete storage pool ID %s Error: %v", pool.Id, err)
		}
	}
}
