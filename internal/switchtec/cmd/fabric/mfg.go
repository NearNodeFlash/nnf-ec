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

package cmd

import (
	"fmt"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
)

// GetSerialCmd -
type GetSerialCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will run the GetSerialCmd command
func (cmd *GetSerialCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		serial, err := dev.GetSerialNumber()

		fmt.Printf("Device Serial Number: %#08x\n", serial)
		return err
	})
}

// GetFirmwareCmd -
type GetFirmwareCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will run the GetFirmwareCmd command
func (cmd *GetFirmwareCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		ver, err := dev.GetFirmwareVersion()

		fmt.Printf("Device Firmware: %s\n", ver)

		return err
	})
}

type MfgCmd struct {
	Serial   GetSerialCmd   `kong:"cmd,help='Retrieve the device serial number.'"`
	Firmware GetFirmwareCmd `kong:"cmd,help='Retrieve the active firmware.'"`
}
