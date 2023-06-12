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

loopCount=0
# while ((loopCount < 1));
while true;
do
    # Power cycle everything
    ./powercycle.sh -t x9000c1j*b*n* x9000c1r[4-7]b0

    # Capture switchtec logs and look for switch hangs
    ../../tools/rabbit-s.sh enable-logging x9000c1j4b0
    ../../tools/rabbit-s.sh clear-logs x9000c1j4b0
    ../../tools/rabbit-s.sh switch-hang-logs x9000c1j4b0

    # Start up nnf-ec
    ssh x9000c1j7b0n0 "./nnf-ec -initializeAndExit" > nnf-ec.log 2>&1
    ssh x9000c1j7b0n0 "tools/nvme.sh list" > nvme-list.log

    switchtecDeviceCount=$(ssh x9000c1j7b0n0 "tools/nvme.sh list | wc -l")
    nvmeDeviceCount=$(ssh x9000c1j7b0n0 "ls /dev/nvme* | wc -l")

    if ((switchtecDeviceCount != 16));
    then
        echo "FAILURE: Expecting 16 switchtec devices, see only" "$switchtecDeviceCount" >> results.out
        break
    fi

    if ((nvmeDeviceCount != 21));
    then
        echo "FAILURE: Expecting 21 nvme devices, see only" "$nvmeDeviceCount" >> results.out
        break
    fi

    loopCount=$((loopCount + 1))
    echo "PowerCycle count" "$loopCount" >> results.out
done

exit 1
