package common

import (
	"regexp"
	"testing"
)

func TestRegExp(t *testing.T) {

	re := regexp.MustCompile("/redfish/v1/StorageServices/NNF/StoragePools/(?P<storagePoolId>\\w+)")

	str := "/redfish/v1/StorageServices/NNF/StoragePools/0/AllocatedVolumes/0"
	matches := re.FindStringSubmatch(str)
	if matches == nil {
		t.Errorf("No matches for string %s", str)
	}

	t.Logf("Storage Pool Id: %s", matches[re.SubexpIndex("storagePoolId")])
}
