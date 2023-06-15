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
shopt -s expand_aliases

# Pull in common utility functions
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# shellcheck source="$SCRIPT_DIR"/_util.sh
source "$SCRIPT_DIR"/_util.sh

usage() {
    cat <<EOF
Display Rabbit Supported NVMe Devices and serial numbers
Usage: $0 COMMAND [ARGS...]

Commands:
    devices                             list devices
    details                             list devices with details
EOF
}

DRIVES=()

case $1 in
    devices)
        getDriveList
        for DRIVE in "${DRIVES[@]}";
        do
            printf "%s\n" "$DRIVE"
        done
        ;;
    details)
        getDriveList
        for DRIVE in "${DRIVES[@]}";
        do
            SerialNumber=$(nvme id-ctrl "${DRIVE}" | grep -E "^sn " | awk '{print $3}')
            Mfg=$(nvme id-ctrl "${DRIVE}" | grep -E "^mn " | awk '{print $3}')
            FW=$(nvme id-ctrl "${DRIVE}" | grep -E "^fr " | awk '{print $3}')
            printf "%s\t%s\t%s\t%s\n" "$DRIVE" "$Mfg" "$SerialNumber" "$FW"
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac
