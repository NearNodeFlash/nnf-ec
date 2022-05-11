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

	"github.com/nearnodeflash/nnf-ec/internal/switchtec/pkg/nvme"
)

// DetachNsCmd detaches the given namespace from the given controller;
// de-activates the given namespace's ID. A namespace must be attached
// to a controller before IO commands may be directed to that namespace.
type DetachNsCmd struct {
	Device      string   `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32   `kong:"arg,required,short='n',help='namespace to attach'"`
	Controllers []uint16 `kong:"arg,required,short='c',help='comma-sep controller list'"`
}

// Run will run the Detach Namespace Command
func (cmd *DetachNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.DetachNamespace(cmd.NamespaceID, cmd.Controllers); err != nil {
		return err
	}

	fmt.Printf("Success, detached nsid: %d\n", cmd.NamespaceID)

	return nil
}
