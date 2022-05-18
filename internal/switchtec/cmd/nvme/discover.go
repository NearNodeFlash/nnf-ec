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
	"strings"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/nvme"
)

type DiscoverCmd struct {
	Regexp string `kong:"arg,required,help='regular expression used to filter drives by model number, serial number, or NQN'"`
}

func (cmd *DiscoverCmd) Run() error {
	devices, err := nvme.DeviceList(cmd.Regexp)
	if err != nil {
		return err
	}

	fmt.Println("Device List:")
	fmt.Println(strings.Join(devices, "\n"))

	return nil
}
