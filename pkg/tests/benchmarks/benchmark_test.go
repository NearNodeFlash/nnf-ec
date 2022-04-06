package benchmarks

import (
	"testing"

	ec "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg"
	nnf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-nnf"
	sf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/rfsf/pkg/models"
)

func BenchmarkStorage(b *testing.B) {
	if b.N >= 32 {
		b.Skip("Unsupport NS count")
	}

	c := ec.NewController(ec.NewMockOptions(true))
	defer c.Close()

	if err := c.Init(nil); err != nil {
		b.Fatalf("Failed to start nnf controller")
	}

	ss := nnf.NewDefaultStorageService()
	b.ResetTimer()

	pools := make([]*sf.StoragePoolV150StoragePool, 0)
	for i := 0; i < b.N; i++ {

		sp := &sf.StoragePoolV150StoragePool{
			CapacityBytes: 1024 * 1024,
		}

		if err := ss.StorageServiceIdStoragePoolsPost(ss.Id(), sp); err != nil {
			b.Fatalf("Failed to create storage pool %d Error: %+v", i, err)
		}

		b.Logf("Created storage pool %d ID: %s", i, sp.Id)
		pools = append(pools, sp)
	}

	for i, pool := range pools {
		if err := ss.StorageServiceIdStoragePoolIdDelete(ss.Id(), pool.Id); err != nil {
			b.Fatalf("Failed to delete storage pool %d ID: %s Error: %v", i, pool.Id, err)
		}

		b.Logf("Deleted storage pool ID %s", pool.Id)
	}
}
