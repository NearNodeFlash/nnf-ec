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
set -e
# set -o xtrace

computeNodes=(
    x9000c3s0b0
    x9000c3s0b1
    x9000c3s1b0
    x9000c3s1b1
    x9000c3s2b0
    x9000c3s2b1
    x9000c3s3b0
    x9000c3s3b1
    x9000c3s4b0
    x9000c3s4b1
    x9000c3s5b0
    x9000c3s5b1
    x9000c3s6b0
    x9000c3s6b1
    x9000c3s7b0
    x9000c3s7b1
)

# if (($poweredOffCount < ${#computeNodes[@]})); then
#     printf "Please power off all compute nodes to begin programming. %d nodes are powered off" "$poweredOffCount"
#     cm power status -t node x9000c3s*b*n*
#     exit 1
# fi

BIOS="ex235a.bios-1.6.1_SBIOS-2866_1.6.1_RabbitSupport.ROM"
BIOSVERSION=$(echo $BIOS | sed 's|.ROM||g')
echo $BIOS $BIOSVERSION
# BIOS="ex235a.bios-1.6.0_SBIOS-2831_RabbitS_hotplug_support.ROM"
# BIOS="ex235a.bios-1.6.0_SBIOS-2827_Verbose_tag2.ROM"

for nC in "${computeNodes[@]}";
do
    # printf "Copy %s to compute node %s\n" "$BIOS" "$nC"
    # scp "$BIOS" "$nC":/tmp

    # printf "Flash %s on %s\n" "$BIOS" "$nC"
    # ssh "$nC" "BIOS=$BIOS && echo BIOS=$BIOS && ls -l /tmp/ex* && node_flashbios 0 /tmp/$BIOS"

    printf "Compute node %s BIOS\n" "$nC"
    ssh "$nC" "cat /rwfs/.redfish/bios/bios_version.json.Node0 | jq"
    # ssh "$nC" "curl -u root:initial0 -k -XGET https://x9000c3s0b0/redfish/v1/Systems/Node0 | jq | grep -i bios"
    # ssh "$nC" "cat /var/log/n0/current | grep CRAY_VERSION | awk '{print $2}'"
    # ssh "$nC" "cat /var/log/n0/current | grep CRAY_VERSION | awk '{print $3}'"

    # printf "Compute node %s BIOS Settings\n" "$nC"
    # ssh "$nC" "curl -u root:initial0 -k -XGET https://x9000c3s0b0/redfish/v1/Systems/Node0/Bios/SD | jq"

    # printf "Compute node %s\n" "$nC"
    # ssh "$nC" "ls /dev/nvme*"
done