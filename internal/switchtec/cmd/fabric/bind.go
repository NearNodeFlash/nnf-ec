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
	"strconv"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
)

// BindCmd defines the Bind CLI command and parameters
type BindCmd struct {
	Device    string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	SwIndex   uint8  `arg help:"Software index of the fabric device."`
	PhyPortID uint8  `arg help:"Physical port ID within the domain."`
	LogPortID uint8  `arg help:"Logical port ID within the domain."`
	PDFID     string `arg help:"PDFID of the end-point."`
}

// Run will run the Bind command
func (cmd *BindCmd) Run() error {
	pdfid, err := strconv.ParseUint(cmd.PDFID, 0, 16)
	if err != nil {
		return err
	}

	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	return dev.Bind(cmd.SwIndex, cmd.PhyPortID, cmd.LogPortID, uint16(pdfid))
}

// UnbindCmd defines the Bind CLI command and parameters
type UnbindCmd struct {
	Device    string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	SwIndex   uint8  `arg help:"Software index of the fabric device."`
	PhyPortID uint8  `arg help:"Physical port ID within the domain."`
	LogPortID uint8  `arg help:"Logical port ID within the domain."`
}

// Run will run the Unbind command
func (cmd *UnbindCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}

	return dev.Unbind(cmd.SwIndex, cmd.PhyPortID, cmd.LogPortID)
}
