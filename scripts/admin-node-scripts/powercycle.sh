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

Usage: $0 [-h] [-t] [RABBIT-P-X-NAMES] [RABBIT-S-X-NAMES]

X-NAMES:
    # Texas TDS systems
    x9000c[0-7]rbt7b0n0             Chassis 0..7, Rabbit P (rbt7), board 0, node 0
    x9000c[0-7]rbt4b0               Chassis 0..7, Rabbit S (rbt4), board 0

    # Stoneship TDS systems
    x1000c[0-7]j7b0n0               Chassis 0..7, Rabbit P (j7), board 0, node 0
    x1000c[0-7]j4b0                 Chassis 0..7, Rabbit S (j4), board 0

Arguments:
  -h                display this help
  -t                time each command

Examples:
  # Texas TDS
  ./powercycle.sh -t x9000c[1,3]rbt7b0n0 x9000c[1,3]rbt4b0              # c[1] - tx-peter, c[3] - tx-bugs

  # Stoneship TDS
  ./powercycle.sh -t x1000c[0-7]j7b0n0 x1000c[0-7]j4b0                  # c[0-7] all Rabbits
EOF
}


rabbitPXName="${1:-x9000c3rbt7b0n0}"
rabbitSXName="${2:-x9000c3rbt4b0}"

paxControl() {
    local op=$1

    if [ -z "$op" ];
    then
        printf "paxControl operation not set\n"
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

    printf "Pax power: %s\n" "$op"
    pdsh -w "$rabbitSXName" pax-power "$op"
}

rabbitPControl() {
    local op=$1

    if [ -z "$op" ];
    then
        printf "RabbitPControl operation not set\n"
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

    printf "RabbitP power: %s\n" "$op"
    cm power "$op" -t node "$rabbitPXName"
}

waitRabbitState() {
    local desiredState=$1
    if [ -z "$desiredState" ];
    then
        printf "waitRabbitState desiredState not set\n"
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
            printf "Invalid Rabbit power state: %s\n" "$desiredState"
            exit 1
            ;;
    esac

    local areWeThereYet="$(cm power status -t node $rabbitPXName | grep -v $desiredState)"
    for ((i=0; ${#areWeThereYet} > 0; i++));
    do
        if (( i > 300 ));
        then
            printf "Waited too long\n"
            exit 1
        fi

        sleep 1
        areWeThereYet="$(cm power status -t node $rabbitPXName | grep -v $desiredState)"

        if ((i % 10 == 0));
        then
            date
            printf "Still waiting for:\n"
            printf "%s\n" "$areWeThereYet"
        fi
    done

    printf "Everything %s\n" "$desiredState"
}

rabbitPControl "Off"
paxControl "Off"
waitRabbitState "Off"

paxControl "On"
rabbitPControl "On"
waitRabbitState "PoweringOn"

waitRabbitState "BOOTED"





# pdsh -w x1000c[0-7]rbt4b0 pax-power off | grep down

# pdsh -w x1000c[0-7]rbt4b0 pax-power on

# RabbitP power cycling:
# cm power status -t node x1000c[0-7]rbt7b0n0

# cm power off -t node x1000c[0-7]rbt7b0n0

# cm power status -t node x1000c[0-7]rbt7b0n0

# cm power on -t node x1000c[0-7]rbt7b0n0

# watch 'cm power status -t node x1000c[0-7]rbt7b0n0'