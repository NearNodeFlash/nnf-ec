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

usage() {
    cat <<EOF
Collect the set of Non Volatile (NV) logs from each PAX switch.
The log files will be stored in the current directory.
Usage: $0 [-h]

Arguments:
  -h                display this help

Example:
  ./$0

  Produces a list like this:
ll *.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-FLASH.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-MEMLOG.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-NVHDR.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-REGS.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-SYS_STACK.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-THRDS.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax0-THRD_STACK.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-FLASH.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-MEMLOG.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-NVHDR.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-REGS.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-SYS_STACK.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-THRDS.log
-rw-r--r-- 1 root root 0 Apr 26 15:36 pax1-THRD_STACK.log

EOF
}

while getopts "h" OPTION
do
    case "${OPTION}" in
        'h')
            usage
            exit 0
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND - 1))


logSources=("FLASH" "MEMLOG" "REGS" "THRD_STACK" "SYS_STACK" "THRDS" "NVHDR")
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