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

// IdCtrlCmd shows secondary controller list associated with the
// primary controller of the given device.
type IdCtrlCmd struct {
	Device string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
}

// Run will run the Identify Controller Command.
func (cmd *IdCtrlCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	ctrl, err := dev.IdentifyController()
	if err != nil {
		return err
	}

	fmt.Printf("Identify Controller:")
	fmt.Printf("%+v", ctrl)

	return nil
}

type IdNamespaceCtrls struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceId uint32 `kong:"arg,required,short='n',help='Namespace to identify.'"`
}

func (cmd *IdNamespaceCtrls) Run() error {
	return run(cmd.Device, func(dev *nvme.Device) error {
		ctrlList, err := dev.IdentifyNamespaceControllerList(cmd.NamespaceId)
		if err != nil {
			return err
		}

		fmt.Printf("Controller List for Namespace ID %d:\n", cmd.NamespaceId)
		fmt.Printf("%+v", ctrlList)

		return nil
	})
}
