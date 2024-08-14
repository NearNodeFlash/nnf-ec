#!/bin/bash

# Copyright 2020-2024 Hewlett Packard Enterprise Development LP
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
  enable-logging        turn on pax logging via screen sessions
  clear-logs            clear current pax logs
  get-logs              retrieve pax logs into current folder
  fabdbg-on             turn on verbose fabric debug
  fabdbg-off            turn off verbose fabric debug
  additional-logs       turn on additional logs
  additional-logs-off   turn off additional logs
  slow-drive-logs       turn on logs for slow drives (0x82827 response code)
  switch-hang-logs      turn on logs to find switch hang
  pcie-error-logs       turn on logs to look for info about pcie fatal error
  quit-sessions         terminate any active screen sessions
  lnkstat               retrieve link state

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
        *)
            usage
            exit 1
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
            PAX=$(echo $DEVICE | cut -d' ' -f1)

            echo "Enabling Logging on $PAX"
            $SSHPASS ssh -T root@$SYSTEM <<-EOF
            if ! screen -list | grep -q "$PAX"; then
                screen -dmS $DEVICE 230400 &&
                screen -S $PAX -X colon "logfile $PAX.log^M" &&
                screen -S $PAX -X colon "logfile flush 1^M" &&
                screen -S $PAX -X colon "log on^M"
            fi
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
    fabdbg-on)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -s info\nfabdbg -s pax\nfabdbg -s gfms\nfabdbg -s hvm\nfabdbg -s sfm\nfabdbg -s fio\nfabdbg -s rule\n"
EOF
        done
        ;;
    fabdbg-off)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -c info\nfabdbg -c pax\nfabdbg -c gfms\nfabdbg -c hvm\nfabdbg -c sfm\nfabdbg -c fio\nfabdbg -c rule\n"
EOF
        done
        ;;
    additional-logs)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "log -t on\nlog -m 0x82 -s3\nlog -m 0x84 -s3\nlog -m 0x82 -s3 -p on\nlog -m 0x84 -s3 -p on\n"
EOF
        done
        ;;
    additional-logs-off)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "log -p off\nlog -m 0x84 -s3 -p off\nlog -m 0x82 -s3 -p off\nlog -m 0x84 -s3 -p off\n"
EOF
        done
        ;;
    slow-drive-logs)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s gfms\nfabdbg -s hvm\nfabdbg -s sfm\nlog -m 0x84 -s 3 -p on -t on\nlog -m 0x82 -s 3 -p on -t on\n"
EOF
        done
        ;;
    switch-hang-logs)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh root@$SYSTEM <<-EOF
            # screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s gfms\nfabdbg -s hvm\nfabdbg -s sfm\nlog -m 0x84 -s 3 -p on -t on\nlog -m 0x82 -s 3 -p on -t on\n"

            # New and improved settings based on https://customer-jira.microchip.com/browse/HPECRAY-23
            screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s gfms\nfabdbg -s hvm\nfabdbg -s sfm\nlog -m 0x84 -s 5 -p on -t on\nlog -m 0x82 -s 5 -p on -t on\nlog -m 0x83 -s 5 -p on -t on\nlog -m 0x80 -s 5 -p on -t on\n"
EOF
        done
        ;;
    pcie-error-logs)
        for SESSION in "${SESSIONS[@]}"
        do
            # per email from Jackson Nguyen:
            # For the PAX logs, please capture the following modules on the UART command line.
            #
            # NVME_MI - 0x81
            # FABIOV - 0x82
            # NVME - 0x84
            # PSC - 0x54
            #
            # These are relevant modules for PAX. Here's an example of activating logs for PSC:
            #
            # > log -m 0x54 -s 5
            # > log -m 0x54 -s 5 -p on
            #
            # You will need to run both commands. You can check the module IDs by typing the `log -l` command.
            # Please also run the following:
            #
            # > fabdbg -s pax
            # > fabdbg -s fio
            # > fabdbg -s gfms

            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s fio\nfabdbg -s gfms\nlog -m 0x81 -s 5\nlog -m 0x81 -s 5 -p on\nlog -m 0x82 -s 5\nlog -m 0x82 -s 5 -p on\nlog -m 0x84 -s 5\nlog -m 0x84 -s 5 -p on\nlog -m 0x54 -s 5\nlog -m 0x54 -s 5 -p on\n"
EOF
        done
        ;;
    hpecray-29)
        for SESSION in "${SESSIONS[@]}"
        do

            # For this run, we want to enable some logging settings that we have been leaving off for a while.
            # FABIOV - 0x82
            #     log -m 0x82 -s 3
            #     log -m 0x82 -s 3 -p on
            # PTD - 0x53
            #     log -m 0x53 -s 3
            #     log -m 0x53 -s 3 -p on
            # fabdbg -s pax
            # fabdbg -s fio
            # fabdbg -s gfms

            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s fio\nfabdbg -s gfms\nlog -m 0x82 -s 3\nlog -m 0x82 -s 3 -p on\nlog -m 0x53 -s 3\nlog -m 0x53 -s 3 -p on\n"
EOF
        done
        ;;
    hpecray-32)
        for SESSION in "${SESSIONS[@]}"
        do
            # Enables medium severity and turns on logging for the PSC module
            # log -m 0x54 -s 3
            # log -m 0x54 -s 3 -p on

            # Enables logs for the fabric debug modules
            # fabdbg -s pax
            # fabdbg -s fio
            # fabdbg -s gfms

            # Turn on the logging for all modules
            # log -p on
            $SSHPASS ssh root@$SYSTEM <<-EOF
            screen -S $SESSION -X stuff "fabdbg -s pax\nfabdbg -s fio\nfabdbg -s gfms\nlog -m 0x54 -s 3\nlog -m 0x54 -s 3 -p on\nlog -p on\n"
EOF
        done
        ;;

    quit-sessions)
        for SESSION in "${SESSIONS[@]}"
        do
            $SSHPASS ssh -T root@$SYSTEM "screen -S $SESSION -X quit"
        done
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