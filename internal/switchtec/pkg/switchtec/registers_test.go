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

package switchtec

import (
	"testing"
	"unsafe"
)

func TestRegisterLayout(t *testing.T) {
	// Not sure about byte-packing the registers

	assertEqual := func(a uintptr, b uintptr, t *testing.T) {
		if a != b {
			t.Fatalf("%v ! %v", a, b)
		}
	}

	var regs mrpcRegs
	assertEqual(unsafe.Offsetof(regs.command), 2048, t)
	assertEqual(unsafe.Offsetof(regs.status), 2052, t)
	assertEqual(unsafe.Offsetof(regs.ret), 2056, t)
}
