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

unbootedComputeNodes="$(cm power status -t node x9000c3s*b*n* | grep -v BOOTED | awk '{print $1}' | sed -e 's|:||g')"
echo $unbootedComputeNodes

powerDownSleep=10
for nC in $unbootedComputeNodes;
do
    printf "Powering off ComputeNode %s\n" "$nC"
    cm power off -t node "$nC"

    printf "Sleeping %d seconds to let controller finish power down\n" "$powerDownSleep"
    sleep "$powerDownSleep"
    cm power on -t node "$nC"
done