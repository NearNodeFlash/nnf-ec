#!/bin/bash

# Copyright 2024 Hewlett Packard Enterprise Development LP
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

shopt -s expand_aliases

logSources=("FLASH" "MEMLOG" "REGS" "THRD_STACK" "SYS_STACK" "THRDS" "NVHDR" "RAM")
switches=("/dev/switchtec0" "/dev/switchtec1")

for switch in "${switches[@]}"; do
    PAX_ID=$(switchtec fabric gfms-dump "$switch" | grep "^PAX ID:" | awk '{print $3}')
    if ! (( PAX_ID >= 0 && PAX_ID <= 1 )); then
        echo "$PAX_ID not in range 0-1"
        exit 1
    fi

    for logSource in "${logSources[@]}"; do
        echo pax"$PAX_ID"-"$logSource".log
        switchtec log-dump "$switch" pax"$PAX_ID"-"$logSource".log --type="$logSource"
    done
done