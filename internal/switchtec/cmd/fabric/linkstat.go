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

// LinkStatCmd defines the Link Stat CLI command and parameters
type LinkStatCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
}

// Run will execute the Link Stat CLI Command
func (cmd *LinkStatCmd) Run() error {
	return run(cmd.Device, func(dev *switchtec.Device) error {
		stats, err := dev.LinkStat()
		if err != nil {
			return err
		}

		fmt.Printf("Link Stats\n")
		for _, s := range stats {
			fmt.Printf("Physical Port ID: %d\n", s.PhysPortId)

			fmt.Printf("  %-32s : %v\n", "Link Up", s.LinkUp)
			fmt.Printf("  %-32s : Gen%d\n", "Link Gen", s.LinkGen)
			fmt.Printf("  %-32s : %04x\n", "Link State", uint16(s.LinkState))

			fmt.Printf("  %-32s : x%d\n", "Configured Link Width", s.CfgLinkWidth)
			fmt.Printf("  %-32s : x%d\n", "Negotiated Link Width", s.NegLinkWidth)

			fmt.Printf("  %-32s : %4.2f GBps\n", "Current Link Rate", s.CurLinkRateGBps)
		}

		return nil
	})
}
