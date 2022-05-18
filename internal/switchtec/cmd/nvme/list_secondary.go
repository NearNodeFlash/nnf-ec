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
	"fmt"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/nvme"
)

// ListSecondaryCmd shows secondary controller list associated with the
// primary controller of the given device.
type ListSecondaryCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	CntID       uint32 `kong:"arg,optional,default='0',help='Lowest controller identifier to display.'"`
	NamespaceID uint32 `kong:"arg,optional,default='0',help='optional namespace attached to controller.'"`
	NumEntries  int    `kong:"arg,optional,default='-1',help='number of entries to display.'"`
}

// Run will run the List Secondary Command
func (cmd *ListSecondaryCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	list, err := dev.ListSecondary(cmd.CntID, cmd.NamespaceID)
	if err != nil {
		return err
	}

	count := list.Count
	if cmd.NumEntries >= 0 {
		count = uint8(cmd.NumEntries)
	}

	fmt.Printf("Identify Secondary Controller List:\n")
	fmt.Printf("  %-12s: %-32s : %d\n", "NUMID", "Number of Identifiers", list.Count)
	for i := uint8(0); i < count; i++ {
		fmt.Printf("  .............\n")
		fmt.Printf("  %-12s:\n", fmt.Sprintf("SCEntry[%3d]", i))

		entry := list.Entries[i]
		fmt.Printf("  %-12s: %-32s : %04x\n", "SCID", "Secondary Controller Identifier", entry.SecondaryControllerID)
		fmt.Printf("  %-12s: %-32s : %04x\n", "PCID", "Primary Controller Identifier", entry.PrimaryControllerID)
		fmt.Printf("  %-12s: %-32s : %04x\n", "SCS", "Secondary Controller State", entry.SecondaryControllerState)
		fmt.Printf("  %-12s: %-32s : %04x\n", "VFN", "Virtual Function Number", entry.VirtualFunctionNumber)
		fmt.Printf("  %-12s: %-32s : %04x\n", "NVQ", "Num VQ Flex Resources Assigned", entry.VQFlexibleResourcesAssigned)
		fmt.Printf("  %-12s: %-32s : %04x\n", "NVI", "Num VI Flex Resources Assigned", entry.VIFlexibleResourcesAssigned)
	}
	if list.Count != count {
		fmt.Printf("  %-12s: %-32s", "", "Display truncated")
	}
	return nil
}
