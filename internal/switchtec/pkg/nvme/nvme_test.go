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

package nvme

import (
	"testing"

	"github.com/HewlettPackard/structex"
)

func TestIdentifyPowerStateStructex(t *testing.T) {
	id := id_power_state{}

	if sz, _ := structex.Size(id); sz != 32 {
		t.Fatal("Power State Size Incorrect")
	}
}

func TestIdentifyControllerStructex(t *testing.T) {
	id := id_ctrl{}

	//id.OptionalAdminCommandSupport.VirtualizationManagment = 1
	id.Reserved1024[0] = 0xFF
	id.VendorSpecific[0] = 0xFF

	buf := structex.NewBuffer(id)
	if err := structex.Encode(buf, id); err != nil {
		t.Fatal(err)
	}

	if buf.Bytes()[1024] != 0xFF {
		t.Fatal("buffer packing failed 1024")
	}
	if buf.Bytes()[3072] != 0xFF {
		t.Fatal("buffer packing failed 3072")
	}

	if len(buf.Bytes()) != 4096 {
		t.Fatalf("Buffer size incorrect: Is %d Expected: %d", len(buf.Bytes()), 4096)
	}

}

func TestAdminCmdStructex(t *testing.T) {
	cmd := AdminCmd{}

	sz, err := structex.Size(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if sz != 72 {
		t.Fatalf("Admin Command Size incorrect: Expected: %d Actual: %d", 72, sz)
	}
}

func TestIdentifyNamespaceStructex(t *testing.T) {

	// structex annotation for legibility preferred
	id := IdNs{}
	id.Size = 1048576
	id.Capacity = 1048576
	id.MultiPathIOSharingCapabilities.Sharing = 1
	id.Reserved192[0] = 0xFF

	buf := structex.NewBuffer(id)
	if err := structex.Encode(buf, id); err != nil {
		t.Fatal(err)
	}

	buf.DebugDump()

	if len(buf.Bytes()) != 4096 {
		t.Fatalf("Encoded id ns buffer is wrong size: Expected 4096 Actual: %d", len(buf.Bytes()))
	}

	if buf.Bytes()[192] != 0xFF {
		t.Fatalf("buffer packing failed")
	}
}
