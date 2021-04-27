package nvme

import (
	"testing"
)

func _TestNsid(t *testing.T) {

	ns := "/dev/nvme0n5"

	id, err := GetNamespaceId(ns)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Namespace %s ID %d", ns, id)
}
