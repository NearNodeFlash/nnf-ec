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

// DeleteNsCmd delete the given namespace by sending a namespace management
// command to the provided device. All controllers should be detached from
// the namespace prior to namespace deletion. A namespace ID becomes inactive
// when that namespace is detached or, if the namespace is not already inactive,
// once deleted.
type DeleteNsCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceID uint32 `kong:"arg,required,short='n',help='namespace to delete'"`
}

// Run will run the Delete Namespace Command
func (cmd *DeleteNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.DeleteNamespace(cmd.NamespaceID); err != nil {
		return err
	}

	fmt.Printf("Success, deleted nsid: %d\n", cmd.NamespaceID)

	return nil
}
