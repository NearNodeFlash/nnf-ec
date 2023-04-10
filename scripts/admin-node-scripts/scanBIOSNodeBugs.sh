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

computeNodesNodeControllers=(
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

for nC in "${computeNodesNodeControllers[@]}";
do
    printf "Compute node %s BIOS:\n" "$nC"
    curl -s -u root:initial0 -k -XGET https://"$nC"/redfish/v1/Systems/Node0 | jq | grep -i bios

    printf "\n"
done
