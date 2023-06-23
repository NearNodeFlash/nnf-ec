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

shopt -s expand_aliases

rabbitP="x1002c0j7b0n0"
rabbitS="x1002c0j4b0"
rabbitSlots="x1002c0r[4-7]b0"
expectedNVMEDeviceCount=18

loopCount=0
# while ((loopCount < 1));
while true;
do
    # Power cycle everything
    ./powercycle.sh -t "$rabbitP" "$rabbitSlots"

    # Capture switchtec logs and look for switch hangs
    ../../tools/rabbit-s.sh enable-logging "$rabbitS"
    ../../tools/rabbit-s.sh clear-logs "$rabbitS"
    ../../tools/rabbit-s.sh switch-hang-logs "$rabbitS"

    # Start up nnf-ec
    ssh "$rabbitP" "chmod +x tools/*.sh"
    ssh "$rabbitP" "./nnf-ec -initializeAndExit" > nnf-ec.log 2>&1
    ssh "$rabbitP" "tools/nvme.sh list" > nvme-list.log
    ssh "$rabbitP" "tools/switch.sh status" > switch-status.log

    missingDriveCount=$(< switch-status.log grep -e Drive | grep -c DOWN)
    switchtecDeviceCount=$(< nvme-list.log wc -l)
    nvmeDeviceCount=$(ssh "$rabbitP" "ls /dev/nvme* | wc -l")

    if ((missingDriveCount > 2));
    then
        echo "FAILURE: Expecting 2 missing drives, see" "$missingDriveCount" >> results.out
        exit 1
    fi

    if ((switchtecDeviceCount != 16));
    then
        echo "FAILURE: Expecting 16 switchtec devices, see" "$switchtecDeviceCount" >> results.out
        exit 1
    fi

    if ((nvmeDeviceCount != expectedNVMEDeviceCount));
    then
        echo "FAILURE: Expecting $expectedNVMEDeviceCount nvme devices, see" "$nvmeDeviceCount" >> results.out
        exit 1
    fi

    loopCount=$((loopCount + 1))
    echo "PowerCycle count" "$loopCount" >> results.out
done
