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

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

// VirtualMgmtCmd describes the Virtualization Management command supported by primary controllers
// that support the Virtualization Enhancements capability.
//
// This command is used for:
//  1. Modifying Flexible Resource allocation for the primary controller
//  2. Assigning Flexible Resources for secondary controllers
//  3. Setting the Online and Offline state for secondary controllers
type VirtualMgmtCmd struct {
	Device       string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	Controller   uint16 `kong:"arg,required,help='controller id'"`
	Action       string `kong:"arg,required,enum='assign,online,offline',help='action to take; one of ${enum}'"`
	ResourceType string `kong:"arg,optional,enum=',vq,vi',help='resource type; one of ${enum}'"`
	NumResources uint32 `kong:"arg,optional,default='0',help='number of resources'"`
}

var (
	actionMap = map[string]nvme.VirtualManagementAction{
		"assign":  nvme.SecondaryAssignAction,
		"online":  nvme.SecondaryOnlineAction,
		"offline": nvme.SecondaryOfflineAction,
	}

	resourceMap = map[string]nvme.VirtualManagementResourceType{
		"vi": nvme.VIResourceType,
		"vq": nvme.VQResourceType,
	}
)

// AfterApply will be called after assignment; used here to require arguments for 'assign' action.
func (cmd *VirtualMgmtCmd) AfterApply() error {
	if actionMap[cmd.Action] == nvme.SecondaryAssignAction {
		if cmd.ResourceType == "" {
			return fmt.Errorf("Secondary assignment requires resource type")
		}
		if cmd.NumResources == 0 {
			return fmt.Errorf("Secondary assignment requires non-zero resources")
		}
	}
	return nil
}

// Run will run the Virtualization Management Command
func (cmd *VirtualMgmtCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.VirtualMgmt(cmd.Controller, actionMap[cmd.Action], resourceMap[cmd.ResourceType], cmd.NumResources); err != nil {
		return err
	}

	fmt.Printf("Virtualization Management Command complete.\n")

	return nil
}
