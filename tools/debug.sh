#!/bin/bash
#
# Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
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

usage() {
    cat <<EOF
Debug commands for analyzing NNF systems
Usage: $0 COMMAND [ARGUMENTS]

Commands:
    switchtec-logs [SYSTEM]        capture switchtec logs from a remote system
EOF
}

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

case $1 in
    switchtec-logs)
        if [ $# -lt 2 ]; then
            usage
            exit 1
        fi

        SYSTEM=$2
        
        declare -a SWITCHES=("switchtec0" "switchtec1")
        for SWITCH in ${SWITCHES[@]};
        do
            echo "$SWITCH Capturing Logs..."
            ssh root@$SYSTEM <<-EOF
            switchtec log-dump /dev/$SWITCH assert_nvhdr.map -t NVHDR && \
            switchtec log-dump /dev/$SWITCH assert_flash.log -t FLASH && \
            switchtec log-dump /dev/$SWITCH assert_ram.log -t RAM && \
            switchtec log-dump /dev/$SWITCH assert_memlog.log -t MEMLOG && \
            switchtec log-dump /dev/$SWITCH assert_regs.log -t REGS && \
            switchtec log-dump /dev/$SWITCH assert_thrd_stack.log -t THRD_STACK && \
            switchtec log-dump /dev/$SWITCH assert_sys_stack.log -t SYS_STACK && \
            switchtec log-dump /dev/$SWITCH assert_thrds.log -t THRDS
EOF
            
            echo "$SWITCH Zipping Logs..."
            ssh root@$SYSTEM "zip $SWITCH.zip *.map *.log"
        
            echo "$SWITCH Retrieving Zip..."
            scp root@$SYSTEM:~/$SWITCH.zip ./
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac