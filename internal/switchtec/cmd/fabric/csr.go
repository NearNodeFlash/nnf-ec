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
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
)

type CsrReadCmd struct {
	Device string `arg:"--device" help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	PDFID  string `arg:"--pdfid" help:"PDFID of the end-point."`
	Addr   string `arg:"--addr" help:"Address to read."`
	Bytes  uint8  `arg:"--bytes" default:"4" help:"Number of bytes to read."`
}

func (cmd *CsrReadCmd) Run() error {
	pdfid, err := strconv.ParseUint(cmd.PDFID, 0, 16)
	if err != nil {
		return err
	}

	addr, err := strconv.ParseUint(cmd.Addr, 0, 16)
	if err != nil {
		return err
	}

	return run(cmd.Device, func(d *switchtec.Device) error {

		data, err := d.CsrRead(uint16(pdfid), uint16(addr), cmd.Bytes)
		if err != nil {
			return err
		}

		var value uint32 = 0
		switch cmd.Bytes {
		case 1:
			value = uint32(data[0])
		case 2:
			value = uint32(binary.LittleEndian.Uint16(data))
		case 4:
			value = binary.LittleEndian.Uint32(data)
		}

		fmt.Printf("%06x - %#0*x\n", addr, cmd.Bytes*2, value)

		return nil
	})
}

type CsrWriteCmd struct {
	Device string `arg:"--device" help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	PDFID  string `arg:"--pdfid" help:"PDFID of the end-point."`
	Addr   string `arg:"--addr" help:"Address to read."`
	Data   string `arg:"--data" help:"Value to write."`
	Bytes  uint8  `arg:"--bytes" default:"4" help:"Number of bytes to read."`
}

func (cmd *CsrWriteCmd) Run() error {
	pdfid, err := strconv.ParseUint(cmd.PDFID, 0, 16)
	if err != nil {
		return err
	}

	addr, err := strconv.ParseUint(cmd.Addr, 0, 16)
	if err != nil {
		return err
	}

	data, err := strconv.ParseUint(cmd.Data, 0, 32)
	if err != nil {
		return err
	}

	return run(cmd.Device, func(d *switchtec.Device) error {
		payload := make([]byte, cmd.Bytes)
		binary.PutUvarint(payload, data)

		return d.CsrWrite(uint16(pdfid), uint16(addr), payload)
	})
}

type CsrCmd struct {
	Read  CsrReadCmd  `cmd:"csrread" help:"Read bytes from a device's Configuration Status Registers."`
	Write CsrWriteCmd `cmd:"csrwrite" help:"Write bytes to a device's Configuration Status Registers."`
}
