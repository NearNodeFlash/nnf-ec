/*
 * Copyright 2020-2025 Hewlett Packard Enterprise Development LP
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
	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
)

// EchoCmd defines the Echo CLI command and parameters
type EchoCmd struct {
	Device  string `arg:"--device" help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Payload uint32 `arg:"--payload" default:"0" help:"The echo payload. The bit-inverse will be returned by the device."`
}

// Run will run the Echo command
func (cmd *EchoCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return dev.Echo(cmd.Payload)
}
