#!/bin/bash

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
Run various commands to configure rabbit-s.
Usage: $0 [-h] [-p password] COMMAND SYSTEM

Commands:
  enable-logging    turn on pax logging via screen sessions
  clear-logs        clear current pax logs
  get-logs          retrieve pax logs into current folder
  lnkstat           retrieve link state (EXPERIMENTAL)

Arguments:
  -p                password for SYSTEM
  -h                display this help

Examples:
  rabbit-s.sh enable-logging rabbit-node-0-s
EOF
}

SSHPASS=""

while getopts "p:h" OPTION
do
    case "${OPTION}" in
        'p')
            SSHPASS="sshpass -p ${OPTARG}"
            ;;
        'h')
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND - 1))

if [ $# -le 1 ]; then
    usage
    exit 1
fi

CMD=${1}
SYSTEM=${2}

declare -a SESSIONS=("pax0" "pax1")

case $CMD in
    enable-logging)
        declare -a DEVICES=("pax0 /dev/ttyS9" "pax1 /dev/ttyS11")
        for DEVICE in "${DEVICES[@]}"
        do
            PAX=$(echo $DEVICE | cut -w -f1)

            echo "Enabling Logging on $PAX"
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -dmS $DEVICE 230400 &&
            screen -S $PAX -X colon "logfile $PAX.log^M" &&
            screen -S $PAX -X colon "logfile flush 1^M" &&
            screen -S $PAX -X colon "log on^M"
EOF
        done
        ;;
    clear-logs)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM "> $SESSION.log"
        done
        ;;
    get-logs)
        $SSHPASS scp root@$SYSTEM:~/*.log ./ 
        ;;
    lnkstat)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X colon "wrap off^M" &&
            screen -S $SESSION -X stuff "lnkstat\\n" &&
            sleep 1 &&
            screen -S $SESSION -X hardcopy &&
            cat hardcopy.0 && rm hardcopy.0
EOF
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac