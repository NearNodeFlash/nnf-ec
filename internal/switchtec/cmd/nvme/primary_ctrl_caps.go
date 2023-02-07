/*
 * Copyright 2023 Hewlett Packard Enterprise Development LP
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

// PrimaryCtrlCapsCmd Send NVMe Identify Primary Controller Capabilities
type PrimaryCtrlCapsCmd struct {
	Device       string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	ControllerId uint16 `kong:"arg,optional,short='c',default='0',help='Controller ID'"`
}

// Run will run the Primary Controller Capabilities command
func (cmd *PrimaryCtrlCapsCmd) Run() error {

	return run(cmd.Device, func(dev *nvme.Device) error {
		caps, err := dev.IdentifyPrimaryControllerCapabilities(cmd.ControllerId)
		if err != nil {
			return err
		}

		fmt.Printf("Identify Primary Controller Capabilities:\n")
		fmt.Printf("%-6s: %-42s : %#x\n", "CNTLID", "Controller Identifier", caps.ControllerId)
		fmt.Printf("%-6s: %-42s : %#x\n", "PORTID", "Port Identifier", caps.PortId)
		fmt.Printf("%-6s: %-42s : %#x\n", "CRT", "Controller Resource Type", caps.ControllerResourceType)

		fmt.Printf("%-6s: %-42s : %d\n", "VQFRT", "VQ Resources Flexible Total", caps.VQResourcesFlexibleTotal)
		fmt.Printf("%-6s: %-42s : %d\n", "VQRFA", "VQ Resources Flexible Assigned", caps.VQResourcesFlexibleAssigned)
		fmt.Printf("%-6s: %-42s : %d\n", "VQRFAP", "VQ Resources Flexible Allocated to Primary", caps.VQResourcesFlexibleAllocatedToPrimary)
		fmt.Printf("%-6s: %-42s : %d\n", "VQPRT", "VQ Resources Private Total", caps.VQResourcesPrivateTotal)
		fmt.Printf("%-6s: %-42s : %d\n", "VQFRSM", "VQ Resources Flexible Secondary Maximum", caps.VQResourcesFlexibleSecondaryMaximum)
		fmt.Printf("%-6s: %-42s : %d\n", "VQGRAN", "VQ Flexible Resource Preferred Granularity", caps.VQFlexibleResourcePreferredGranularity)

		fmt.Printf("%-6s: %-42s : %d\n", "VIFRT", "VI Resources Flexible Total", caps.VIResourcesFlexibleTotal)
		fmt.Printf("%-6s: %-42s : %d\n", "VIRFA", "VI Resources Flexible Assigned", caps.VIResourcesFlexibleAssigned)
		fmt.Printf("%-6s: %-42s : %d\n", "VIRFAP", "VI Resources Flexible Allocated to Primary", caps.VIResourcesFlexibleAllocatedToPrimary)
		fmt.Printf("%-6s: %-42s : %d\n", "VIPRT", "VI Resources Private Total", caps.VIResourcesPrivateTotal)
		fmt.Printf("%-6s: %-42s : %d\n", "VIFRSM", "VI Resources Flexible Secondary Maximum", caps.VIResourcesFlexibleSecondaryMaximum)
		fmt.Printf("%-6s: %-42s : %d\n", "VIGRAN", "VI Flexible Resource Preferred Granularity", caps.VIFlexibleResourcePreferredGranularity)

		return nil

	})
}
