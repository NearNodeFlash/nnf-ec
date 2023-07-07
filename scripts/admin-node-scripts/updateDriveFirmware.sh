#!/bin/bash

# Copyright 2023 Hewlett Packard Enterprise Development LP
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
# set -e
# set -o xtrace

usage() {
    cat <<EOF
Query drive firmware version for all drives in a Rabbit. Update drives that are out-of-date.

Assumes that the following are installed on the Rabbit:
- /root/nnf-ec
- /root/tools/nvme.sh
- <firmware-file-path>

Usage: $0 [-h] [-t] [RABBIT-XNAME] [EXPECTED-FIRMWARE] [FIRMWARE-FILE-PATH]

Arguments:
  -h                display this help
  -t                time the operation

Examples:
    ./updateDriveFirmware.sh -h                                                     # Display help message
    ./updateDriveFirmware.sh x9000c3j7b0n0 1TCRS103 /root/KIOXIA/1TCRS103.std     # Rabbit: x9000c3j7b0n0, Expected Firmware: "1TCRS103", Firmware File Path: "x9000c3j7b0n0:/root/KIOXIA/1TCRS103.std"

    ./updateDriveFirmware.sh root@10.1.1.5 1TCRS103 /root/KIOXIA/1TCRS103.std
EOF
}

alias TIME=""
while getopts "th" OPTION;
do
    case "${OPTION}" in
        t)
            alias TIME=time
            export TIMEFORMAT='%3lR'
            ;;
        h)
            usage
            exit 0
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND - 1))

if [ $# -lt 3 ]; then
    usage
    exit 1
fi

rabbit=$1
expectedFirmware=$2
firmwareFile=$3

printf "Validating firmware %s on Rabbit %s\n" "$expectedFirmware" "$rabbit"

# Run nnf-ec to initialize PAX chips and drives
printf "Initialize PCIe Switch connections to drives\n"
result=$(ssh "$rabbit" ./nnf-ec -initializeAndExit >/dev/null)

if [ $? -ne 0 ]
then
    printf "%s\n" "$result"
    exit 1
fi

# Retrieve a list of unique firmware levels
firmware=$(ssh "$rabbit" "tools/nvme.sh cmd id-ctrl | grep -e \"^fr \" | uniq")
firmware=$(echo "$firmware" | awk '{print $3}')
echo "$firmware"

if [ "$firmware" == "$expectedFirmware" ]; then
    printf "Firmware up to date\n"
else
    printf "Firmware mismatch, downloading %s %s\n" "$expectedFirmware" "$firmwareFile"

    for (( slot=1; slot <= 3; ++slot ));
    do
        # shellcheck disable=SC2029
        ssh "$rabbit" "tools/nvme.sh cmd fw-download --fw=$firmwareFile --xfer=256"
        # shellcheck disable=SC2029
        ssh "$rabbit" "tools/nvme.sh cmd fw-activate --slot=$slot --action=1"
    done
fi
printf "\n"
