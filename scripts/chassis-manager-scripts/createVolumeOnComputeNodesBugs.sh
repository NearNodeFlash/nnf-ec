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

computeNodes=(
    x9000c3s0b0n0
    x9000c3s0b1n0
    x9000c3s1b0n0
    x9000c3s1b1n0
    x9000c3s2b0n0
    x9000c3s2b1n0
    x9000c3s3b0n0
    x9000c3s3b1n0
    x9000c3s4b0n0
    x9000c3s4b1n0
    x9000c3s5b0n0
    x9000c3s5b1n0
    x9000c3s6b0n0
    x9000c3s6b1n0
    x9000c3s7b0n0
    x9000c3s7b1n0
)

for nC in "${computeNodes[@]}";
do
    printf "Compute node %s:\n" "$nC"
    rsync lvm.sh "$nC":

    rabbitVGS=$(ssh "$nC" 'vgs | grep rabbit')
    if [ -z "${rabbitVGS[@]}" ];
    then
        # Retrieve the list of namespaces from each KIOXIA or 'SAMSUNG MZ3LO1T9HCJR' drive seen on the compute node and count this number of namespaces. We expect only 1
        nameSpaceCount=$(ssh "$nC" 'for DRIVE in $(ls -v /dev/nvme* | grep -E "nvme[[:digit:]]+n[[:digit:]]+$"); do if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ]; then nvme id-ns $DRIVE | grep "NVME"; fi; done | uniq | wc -l')
        if ((nameSpaceCount > 1)); then
            printf "Too many namespaces(%d), please examine your setup\n" "$nameSpaceCount"
            exit 1
        fi

        # Pull the namespace list from each KIOXIA or 'SAMSUNG MZ3LO1T9HCJR' drive, we know there is only 1 now.
        nameSpaceStr=$(ssh "$nC" 'for DRIVE in $(ls -v /dev/nvme* | grep -E "nvme[[:digit:]]+n[[:digit:]]+$"); do if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ]; then nvme id-ns $DRIVE | grep "NVME"; fi; done | uniq')
        nameSpaceID=$(echo $nameSpaceStr | sed 's|:||g' | awk '{print $4}')

        # Create an LVM volume from the namespaces present
        ssh "$nC" "./lvm.sh create rabbit $nameSpaceID"
    fi
done
printf "\n"


