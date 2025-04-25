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
	"fmt"

	"github.com/HewlettPackard/structex"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
)

type DumpEpPortCmd struct {
	Device string `arg:"--device" help:"The switchtec device." type:"existingFile" env:"SWITCHTEC_DEV"`
	Pid    uint8  `arg:"--pid" help:"The end-point port ID."`
}

func (cmd *DumpEpPortCmd) Run() error {

	fmt.Printf("%s Dump EP Port %d\n", cmd.Device, cmd.Pid)

	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	fmt.Printf("PAX %d Device Opened. Port %d Start...\n", dev.ID(), cmd.Pid)

	len, err := dev.GfmsEpPortStart(cmd.Pid)
	if err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Started. Port Get %d dwords...\n", len)

	buf, err := dev.GfmsEpPortGet(cmd.Pid, len)
	if err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Buf %d bytes \n%v\n", buf.Len(), buf)
	fmt.Printf("Port Finish...\n")

	if err := dev.GfmsEpPortFinish(); err != nil {
		return err
	}

	fmt.Printf("Dump EP Port Finished\n")

	ep := new(switchtec.DumpEpPortDevice)

	if true {
		for i := 0; i < buf.Len(); i++ {
			if i%16 == 0 {
				fmt.Printf("\n%08x: ", i)
			}
			if i%8 == 0 {
				fmt.Printf(" ")
			}
			fmt.Printf("%02x", buf.Bytes()[i])
		}

		fmt.Printf("\n")
	}

	if err := structex.DecodeByteBuffer(buf, ep); err != nil {
		return err
	}

	fmt.Printf("%+v\n", ep)

	return nil
}

type DumpCmd struct {
	EpPort DumpEpPortCmd `cmd:"epport" help:"Dump information for a specific end-point port."`
}
