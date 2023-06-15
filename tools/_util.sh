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

# Retrieve the Physical Device Fabric IDs used to iterate through a list of nvme drives
function getPDFIDs() {
    local SWITCH=$1 FUNCTION="${2:-0}"

    switchtec fabric gfms-dump "$SWITCH" | grep "Function $FUNCTION " -A2 | grep PDFID | awk '{print $2}'
}

function getDriveList() {
    # DRIVES=$1
    # for DRIVE in $(ls /dev/nvme* | grep -E "nvme[[:digit:]]+$");
    for DRIVE in /dev/nvme[0-9]*;
    do
        # shellcheck disable=SC2086
        if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ];
        then
            # SerialNumber=$(nvme id-ctrl ${DRIVE} | grep -E "^sn " | awk '{print $3}')
            # Mfg=$(nvme id-ctrl ${DRIVE} | grep -E "^mn " | awk '{print $3}')
            # FW=$(nvme id-ctrl ${DRIVE} | grep -E "^fr " | awk '{print $3}')
            # printf "%s\t%s\t%s\t%s\n" "$DRIVE" "$Mfg" "$SerialNumber" "$FW"

            DRIVES+=("${DRIVE}")
        fi
    done

    DriveCount="${#DRIVES[@]}"
    if ((DriveCount == 0));
    then
        printf "No drives found: Did you run nnf-ec?\n"
    fi
}
