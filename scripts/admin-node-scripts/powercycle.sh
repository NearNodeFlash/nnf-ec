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

shopt -s expand_aliases

usage() {
    cat <<EOF
Powercycle the specified Rabbits.
Use 'pdsh' style specifiers to specify multiple Rabbit nodes
See: https://linux.die.net/man/1/pdsh for details

Usage: $0 [-h] [-t] [RABBIT-P-X-NAMES] [RABBIT-SLOT-X-NAMES]

X-NAMES:
    # Texas TDS systems
    # Rabbit-p names
    x9000c[0-7]j7b0n0             Chassis 0..7, Rabbit P (j7), board 0, node 0

    # Slot names for Rabbit-p and Rabbit-s
    x9000c1r[4-7]b0

Arguments:
  -h                display this help
  -t                time each command

Examples:
  # Texas TDS
  ./powercycle.sh -t x9000c1j*b*n* x9000c1r[4-7]b0                  # c[1] tx-peter, r[4-7] slots 4-7 (Rabbit-s -> Rabbit-p)
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

rabbitPXName="${1:-x9000c3j7b0n0}"
rabbitSlotXName="${2:-x9000c3j4b0}"

powerControl() {
    local op=$1
    local nameToPower=$2
    local xnameToPower=$3
    local powerSelector

    if [ -z "$op" ];
    then
        printf "RabbitPControl operation not set\n"
        exit 1
    fi

    if [ -z "$nameToPower" ];
    then
        printf "nameToPower not set\n"
        exit 1
    fi

    if [ -z "$xnameToPower" ];
    then
        printf "xnameToPower not set\n"
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

    case "$nameToPower" in
    rabbit)
        powerSelector="node"
        ;;
    slot)
        powerSelector="slot"
        ;;
    *)
        printf "Invalid power control object: %s\n" "$nameToPower"
        exit 1
        ;;
    esac


    printf "%s power: %s, %s\n" "$nameToPower" "$op" "$xnameToPower"
    cm power "$op" -t "$powerSelector" "$xnameToPower"
}

waitForState() {
    local desiredState=$1
    local whatIsWaiting=$2
    local xnameToWaitFor=$3
    local powerSelector

    if [ -z "$desiredState" ];
    then
        printf "waitForState desiredState not set\n"
        exit 1
    fi

    if [ -z "$xnameToWaitFor" ];
    then
        printf "waitForState xnameToWaitFor not set\n"
        exit 1
    fi

    case "$desiredState" in
        On)
            ;;
        Off)
            ;;
        PoweringOn)
            ;;
        BOOTED)
            ;;
        *)
            printf "Invalid power state: %s\n" "$desiredState"
            exit 1
            ;;
    esac

    case "$whatIsWaiting" in
        rabbit)
            powerSelector="node"
            ;;
        slot)
            powerSelector="slot"
            ;;
        *)
            printf "Invalid waiter: %s\n" "$whatIsWaiting"
            exit 1
            ;;
    esac

    # Retrieve the list of Rabbits that are not yet in their desired state
    local areWeThereYet
    areWeThereYet="$(cm power status -t $powerSelector "$xnameToWaitFor" | grep -v "$desiredState")"
    for ((i=0; ${#areWeThereYet} > 0; i++));
    do
        if (( i > 300 ));
        then
            printf "Waited too long\n"
            exit 1
        fi

        if ((i % 10 == 0));
        then
            date
            printf "Waiting for %s(s) to transition to %s:\n" "$whatIsWaiting" "$desiredState"
            printf "%s\n" "$areWeThereYet"
        fi

        areWeThereYet="$(cm power status -t $powerSelector "$xnameToWaitFor" | grep -v "$desiredState")"
    done

    printf "All %s(s) %s\n" "$whatIsWaiting" "$desiredState"
}

TIME powerControl "Off" "rabbit"        "$rabbitPXName"
TIME waitForState "Off" "rabbit"        "$rabbitPXName"
TIME powerControl "Off" "slot"          "$rabbitSlotXName"
TIME waitForState "Off" "slot"          "$rabbitSlotXName"

TIME powerControl "On"  "slot"          "$rabbitSlotXName"
sleep 10
TIME waitForState "On"  "slot"          "$rabbitSlotXName"

# Wait until we can successfully retrieve Rabbit-p's state which indicates that the node controller is ready
TIME waitForState "Off" "rabbit"        "$rabbitPXName"

TIME powerControl "On" "rabbit"         "$rabbitPXName"
TIME waitForState "PoweringOn" "rabbit" "$rabbitPXName"
TIME waitForState "BOOTED" "rabbit"     "$rabbitPXName"
