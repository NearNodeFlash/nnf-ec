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
Powercycle the specified compute nodes.
Use 'pdsh' style specifiers to specify the compute nodes
See: https://linux.die.net/man/1/pdsh for details

Usage: $0 [-h] [-t] [compute-node-names]

Arguments:
  -h                display this help
  -t                time each command

Examples:
  ./powercycle.sh -t elcap[8681-8808]
EOF
}


alias TIME=""
while getopts "th:" OPTION
do
    case "${OPTION}" in
        't')
            alias TIME=time
            export TIMEFORMAT='%3lR'
            ;;
        'h',*)
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND - 1))

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

nodeName="${1:-elcap[8681-8808]}"

powerControl() {
    local op=$1
    local nameToPower=$2

    if [ -z "$op" ];
    then
        printf "Power Control operation not set\n"
        exit 1
    fi

    if [ -z "$nameToPower" ];
    then
        printf "nameToPower not set\n"
        exit 1
    fi

    case "$op" in
        On)
            op="on"
            ;;
        Off)
            op="off"
            ;;
        *)
            printf "Invalid power cycle option: %s\n" "$op"
            exit 1
            ;;
    esac

    printf "%s power: %s\n" "$nameToPower" "$op"
    pdsh -f 128 -w "p$nameToPower" "redfish node 0 $op"
}

waitForBooted() {

    local nameToWaitFor=$1
    waitingFor="$(pdsh -f 128 -w "e$nameToWaitFor" "tail -1 /var/log/ansible.log" 2>&1 | grep "timed out" | awk '{print $1}' | sed 's|:||g')"
    for (( i=0; ${#waitingFor} > 0; i++ ));
    do
        if ((i > 3000));
        then
            printf "Waited too long\n"
            exit 1
        fi

        if ((i % 10 == 0));
        then
            date
            printf "Waiting for\n"
            printf "%s\n" "$waitingFor"
        fi

        sleep 2
        waitingFor="$(pdsh -f 128 -w "e$nameToWaitFor" "tail -1 /var/log/ansible.log" 2>&1 | grep "timed out" | awk '{print $1}' | sed 's|:||g')"
    done

    printf "All %s\n" "booted"
}

TIME powerControl "Off"         "$nodeName"

sleep 30

TIME powerControl "On"          "$nodeName"

sleep 30

TIME waitForBooted              "$nodeName"
