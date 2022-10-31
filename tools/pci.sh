#!/bin/bash

# Copyright 2022 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

usage() {
    cat << EOF
Shows the hexadecimal dump of the whole PCI configuration space for the
provided switchtec device identifer. The output can be saved to a fil
and then parsed using 'lspci -F <file>'.

Usage: $0 SWITCH PDFID [COUNT=64]

  SWITCH must be a valid Switchtec device (ex: /dev/switchtec0)
  PDFID must be the physical device function ID
  COUNT is the number of 32-bit register reads (default 64)

Example:
  pci.sh /dev/switchtec0 0x1800

  pci.sh /dev/switchtec0 0x1800 > temp.txt && lspci -F temp.txt -vvv
EOF
}

if [ $# -lt 2 ]; then
    usage
    exit 1
fi

SWITCH=$1
PDFID=$2

COUNT=${3:-64}

# Read the PCIe Config Space Registers
# The output is the equivalent of `printf "%06X - %#08X\n" $ADDRESS $VALUE`
IFS=$'\n' read -ra LINES -d $'\0' <<< "$(switchtec fabric ep-csr-read "${SWITCH}" --pdfid="${PDFID}" --addr=0 --bytes=4 --count="${COUNT}" --print=hex)"

# In trying to mimic the `lspci -xx` output which will print the device summary,
# followed by the hexadecimal dump of the standard part of the configuration space.
# For example: lspci -s 05:00.0 -xx
#  05:00.0 Non-Volatile memory controller: KIOXIA Corporation Device 0014 (rev 01)
#  00: 0f 1e 14 00 04 00 10 00 01 02 08 01 00 00 80 00
#  [...]
#  30: 00 00 00 00 70 00 00 00 00 00 00 00 00 00 00 00

# The leading device summary must be present, even if it does not match the data
echo -n "00:00.0 UNKNOWN NVME DEVICE"

for INDEX in "${!LINES[@]}"; do

    # Print the byte offset prefix ever 16 bytes
    if [ $((INDEX%4)) == 0 ]; then
        printf "\n"
        printf "%02x: " $((INDEX*4))
    fi

    # Parse the line, grabbing the 32-bit data portion
    DATA=$(echo "${LINES[$INDEX]}" | cut -d " " -f 3)

    # Print individual bytes (little-endian)
    echo -n "${DATA:8:2} ${DATA:6:2} ${DATA:4:2} ${DATA:2:2} "

done

echo "" # Ending newline
