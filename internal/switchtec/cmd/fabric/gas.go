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
	"encoding/binary"
	"fmt"
	"strconv"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

// GasWriteCmd defines the GAS Write CLI command and parameters
type GasWriteCmd struct {
	Device string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Addr   string `arg help:"Address to write."`
	Value  string `arg help:"Value to write."`
	Bytes  int    `arg optional default:"4" help:"Number of bytes to write."`
}

// Run will execute the GAS Write Command
func (cmd *GasWriteCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}

	addr, err := strconv.ParseUint(cmd.Addr, 0, 32)
	if err != nil {
		return err
	}

	value, err := strconv.ParseUint(cmd.Value, 0, cmd.Bytes*8)
	if err != nil {
		return err
	}

	switch cmd.Bytes {
	case 1:
		err = dev.GASWrite([]byte{uint8(value)}, addr)
	case 2:
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b[:], uint16(value))
		err = dev.GASWrite(b, addr)
	case 4:
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b[:], uint32(value))
		err = dev.GASWrite(b, addr)
	case 8:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b[:], uint64(value))
		err = dev.GASWrite(b, addr)
	default:
		err = fmt.Errorf("Invalid access width")
	}

	return err
}

// GasReadCmd defines the GAS Read CLI command and parameters
type GasReadCmd struct {
	Device string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Addr   string `arg help:"Address to read."`
	Bytes  uint64 `arg optional default:"4" help:"Number of bytes to read."`
}

// Run will execute the GAS Read Command and display the read data
func (cmd *GasReadCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}

	addr, err := strconv.ParseUint(cmd.Addr, 0, 32)
	if err != nil {
		return err
	}

	b, err := dev.GASRead(addr, cmd.Bytes)
	if err != nil {
		return err
	}

	var value uint64 = 0
	switch cmd.Bytes {
	case 1:
		value = uint64(b[0])
	case 2:
		value = uint64(binary.LittleEndian.Uint16(b))
	case 4:
		value = uint64(binary.LittleEndian.Uint32(b))
	case 8:
		value = binary.LittleEndian.Uint64(b)
	default:
		return fmt.Errorf("Invalid access width")
	}

	fmt.Printf("%06x = %#0*x\n", addr, cmd.Bytes*2, value)

	return nil
}

type GasStatCmd struct {
	Device string `arg help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
}

func (cmd *GasStatCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}

	path, err := dev.SystemPath("")
	if err != nil {
		return err
	}

	fmt.Printf("Device Path: %s\n", path)

	size, err := dev.ResourceSize("device/resource0")
	if err != nil {
		return err
	}

	fmt.Printf("Resource0 Size: %d\n", size)

	/*
		if err := dev.GasMap()
		if _, err := dev.NewGAS(true); err != nil {
			return fmt.Errorf("GAS Access denied: %v", err)
		}
	*/

	return nil
}

type GasCmd struct {
	Write GasWriteCmd `cmd help:"Write bytes to the Global Address Space."`
	Read  GasReadCmd  `cmd help:"Read bytes from the Global Address Space."`
	Stat  GasStatCmd  `cmd help:"Identify the device's Global Address Space."`
}
