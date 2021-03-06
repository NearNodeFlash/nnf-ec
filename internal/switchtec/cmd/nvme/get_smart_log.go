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

type GetSmartLogCmd struct {
	Device string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
}

func (cmd *GetSmartLogCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	log, err := dev.GetSmartLog()
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", log)

	return nil
}
