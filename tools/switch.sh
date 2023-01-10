#!/bin/bash

# Copyright 2022 Hewlett Packard Enterprise Development LP
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

shopt -s expand_aliases

usage() {
    cat <<EOF
Run various switch command over all switches.

Usage: $0 [-h] [-t] COMMAND [ARGS...]

Commands:
    slot-info                            display slot status for each physical port
    status                               display port status
    info                                 display switch hardware information
    fabric [COMMAND] [ARG [ARG]...]      execute the fabric COMMAND (default is gfms-dump)

    cmd [COMMAND] [ARG [ARG]...]         execute COMMAND on each switchtec device in the system.
                                         i.e. $0 fw-info

Arguments:
  -h                display this help
  -t                time each command

Examples:
  ./switch.sh slot-info                                                         # display physical slot -> switch physical port status
  ./switch.sh slot-info | grep 'Not attached'                                   # display slots without a working drive

  ./switch.sh status                                                            # display switch port information
  ./switch.sh info                                                              # display switch information

  ./switch.sh fabric                                                            # defaults to 'gfms-dump' command to dump the GFMS database
  ./switch.sh fabric | grep '^PAX ID:' -A4                                      # display the number of endpoints attached to each PAX
  ./switch.sh fabric | grep -e "Physical Port ID " -A2 -e "^PAX ID:"            # display the drive physical ports along with their PDFIDs

  ./switch.sh fabric gfms-dump                                                  # dump GFMS database (same as ./switch.sh fabric)
  ./switch.sh fabric gfms-dump | grep -e "Execute" -e "^PAX ID:"                # display PAX ID associated with each /dev/switchtec device
  ./switch.sh fabric gfms-dump | grep "Function 0 (SRIOV-PF)" -A1 | grep PDFID  # display the list of physical device fabric IDs for the drives attached

  ./switch.sh fabric topo-info                                                  # show topology info including PCIe rates

  ./switch.sh cmd fw-info                                                       # return information on the currently flashed firmware
  ./switch.sh cmd fw-info | grep -A1 'Currently Running'                        # display currently running firmware
EOF
}

# execute <fn<path>> <args...>
execute() {
    local FUNCTION=$1 ARGS=( "${@:2}" )

    if [ "$(type -t "$FUNCTION")" != "function" ]; then
        echo "$1 is not a function."
        exit 1
    fi

    local SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
    for SWITCH in "${SWITCHES[@]}";
    do
        "$FUNCTION" "$SWITCH" "${ARGS[@]}"
    done
}

displaySlotStatus() {
    local SWITCH_NAME=$1
    local PAX_ID

    # Physical slot ids are set into the hardware. These are the mappings
    declare -a PAX0_SlotFromPhysicalPort=(
        [8]=8
        [10]=7
        [12]=15
        [14]=16
        [16]=17
        [18]=18
        [20]=14
        [22]=13
        [48]=12
    )
    declare -a PAX1_SlotFromPhysicalPort=(
        [8]=4
        [10]=5
        [12]=6
        [14]=2
        [16]=1
        [18]=9
        [20]=10
        [22]=11
        [48]=3
    )

    # Make sure we can get the PAX ID
    if [ ! "$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "^PAX ID:" | awk '{print $3}')" ]; then
        exit $?
    fi

    PAX_ID=$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "^PAX ID:" | awk '{print $3}')

    if ! [[ $PAX_ID =~ [[:digit:]]+ ]]; then
        echo "$PAX_ID is not an integer"
        exit 1
    elif ! (( PAX_ID >= 0 && PAX_ID <= 1 )); then
        echo "$PAX_ID not in range 0-1"
        exit 1
    fi

    # Grab the attach status of each physical port
    mapfile -t physicalPortStatus < <(switchtec fabric gfms-dump "$SWITCH" | grep " Physical Port ID")

    local physicalPortString
    for physicalPortString in "${physicalPortStatus[@]}";
    do
        local PHYSICAL_PORT_ID
        PHYSICAL_PORT_ID=$(echo "$physicalPortString" | awk '{print $4}')
        case $PAX_ID in
            0)
                SLOT=${PAX0_SlotFromPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            1)
                SLOT=${PAX1_SlotFromPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            *)
                exit 1
        esac

        printf "DEVICE: %s PAX_ID: %d SLOT: %d\t%s\n" "$SWITCH_NAME" "$PAX_ID" "${SLOT//}" "${physicalPortString//}"
    done
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

case $1 in
    slot-info)
        function slot-info() {
            local SWITCH=$1
            echo "Execute slot-info on $SWITCH"
            TIME displaySlotStatus "$SWITCH"
        }
        execute slot-info
        ;;
    info)
        function info() {
            local SWITCH=$1
            echo "Execute switch info on $SWITCH"
            TIME switchtec info "$SWITCH"
        }
        execute info
        ;;
    status)
        function status() {
            local SWITCH=$1
            echo "Execute switch status on $SWITCH"
            TIME switchtec status "$SWITCH"
        }
        execute status
        ;;
    fabric)
        function fabric() {
            local SWITCH=$1 FABRIC_CMD=$2 ARGS=( "${@:3}" )
            echo "Execute switch fabric $FABRIC_CMD on $SWITCH"
            TIME switchtec fabric "$FABRIC_CMD" "$SWITCH" "${ARGS[@]}"
        }
        execute fabric "${2:-gfms-dump}" "${@:3}"
        ;;
    cmd)
        function cmd() {
            local SWITCH=$1 CMD=$2 ARGS=( "${@:3}" )
            echo "Execute on $SWITCH $CMD" "${ARGS[@]}"
            TIME switchtec "$CMD" "$SWITCH" "${ARGS[@]}"
        }
        execute cmd "${@:2}"
        ;;
    *)
        usage
        exit 1
        ;;
esac
