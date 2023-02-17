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

package switchtec

import "fmt"

const (
	// TODO: Document these registers
	pcieCapabilitiesOffset uint16 = 0x34
	pcieCapabilitiesMask   uint8  = 0xFC

	pcieCapabilitiesId uint8 = 0x10

	pcieDeviceControlerRegisterOffset uint8  = 0x08
	pcieDeviceResetFlag               uint16 = 0x8000

	pcieNextCapabilityOffset uint8 = 0x01
)

func (dev *Device) VfReset(pdfid uint16) error {

	offset, err := dev.CsrRead8(pdfid, pcieCapabilitiesOffset)
	if err != nil {
		return fmt.Errorf("vf reset - failed to read register address %#02x", pcieCapabilitiesOffset)
	}

	offset &= pcieCapabilitiesMask
	for offset != 0 {
		capId, err := dev.CsrRead8(pdfid, uint16(offset))
		if err != nil {
			return fmt.Errorf("vf reset - failed to read register address %#02x", offset)
		}

		if capId == pcieCapabilitiesId {
			break
		}

		offset += pcieNextCapabilityOffset
		nextOffset, err := dev.CsrRead8(pdfid, uint16(offset))
		if err != nil {
			return fmt.Errorf("vf reset - failed to read register address %#02x", offset)
		}

		offset = nextOffset
	}

	if offset == 0 {
		return fmt.Errorf("vf reset - cannot find capability register '%#02x'", pcieCapabilitiesId)
	}

	offset += pcieDeviceControlerRegisterOffset
	ctrl, err := dev.CsrRead16(pdfid, uint16(offset))
	if err != nil {
		return fmt.Errorf("vf reset - failed to read register %#04x", offset)
	}

	ctrl |= pcieDeviceResetFlag
	if err := dev.CsrWrite16(pdfid, uint16(offset), ctrl); err != nil {
		return fmt.Errorf("vf reset - failed to write register %#04x", offset)
	}

	return nil
}
