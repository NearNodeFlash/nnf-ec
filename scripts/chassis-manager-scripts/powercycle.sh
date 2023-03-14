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

rabbitPXName="x1000c[0-7]rbt7b0n0"
rabbitSXName="x1000c[0-7]rbt4b0"

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
    for ((i=0; ${#areWeThereYet} > 300; i++));
    do
        if (( i > 9 ));
        then
            printf "Waited too long\n"
            exit 1
        fi

        sleep 1
        areWeThereYet="$(cm power status -t node $rabbitPXName | grep -v $desiredState)"

        if ((i % 10 == 0));
        then
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