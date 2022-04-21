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
