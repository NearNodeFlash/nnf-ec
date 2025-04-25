/*
 * Copyright 2023-2025 Hewlett Packard Enterprise Development LP
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

type VfResetCmd struct {
	Device string `arg:"--device" help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	PDFID  string `arg:"--pdfid" help:"PDFID of the end-point."`
}

func (cmd *VfResetCmd) Run() error {
	pdfid, err := strconv.ParseUint(cmd.PDFID, 0, 16)
	if err != nil {
		return err
	}

	return run(cmd.Device, func(d *switchtec.Device) error {
		return d.VfReset(uint16(pdfid))
	})
}
