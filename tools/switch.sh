#!/bin/bash

# Copyright 2022-2024 Hewlett Packard Enterprise Development LP
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

# Pull in common utility functions
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# shellcheck source="$SCRIPT_DIR"/_util.sh
source "$SCRIPT_DIR"/_util.sh

usage() {
    cat <<EOF
Run various switch command over all switches.

Usage: $0 [-hv] [-t] COMMAND [ARGS...]

Commands:
    slot-info                            display slot status for each physical port
    status                               display port status for Rabbit connections
    switchtec-status                     display switchtec utility's port status inf
    info                                 display switch hardware information
    ep-tunnel-status                     display endpoint tunnel status for each drive
    ep-tunnel-enable                     enable the endpoint tunnel for each drive
    fabric [COMMAND] [ARG [ARG]...]      execute the fabric COMMAND (default is gfms-dump)

    cmd [COMMAND] [ARG [ARG]...]         execute COMMAND on each switchtec device in the system.
                                         i.e. $0 fw-info

Arguments:
  -h                display this help
  -v                verbose mode
  -t                time each command

Examples:
  ./switch.sh slot-info                                                         # display physical slot -> switch physical port status
  ./switch.sh slot-info | grep 'Not attached'                                   # display slots without a working drive

  ./switch.sh status                                                            # display connected endpoint status
  ./switch.sh switchtec-status                                                  # display switchtec port status
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

# Associative array /dev/nvme names keyed by serial number
declare -A deviceName

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

getChassis() {
    if [ "$VERBOSE" != "true" ]; then
        CHASSIS="       "
        return
    fi

    COMMAND=xhost-query.py
    if command -v $COMMAND &> /dev/null; then
        CHASSIS=$("$COMMAND" $(hostname) | cut -c -7)
    else
        CHASSIS="x****c*"
    fi
}

getPAXID() {
    local SWITCH_NAME=$1

    # Make sure we can get the PAX ID
    if [ ! "$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "^PAX ID:" | awk '{print $3}')" ]; then
        echo "Unable to retrieve PAX ID"
        exit $?
    fi

    PAX_ID=$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "^PAX ID:" | awk '{print $3}')
    if ! (( PAX_ID >= 0 && PAX_ID <= 1 )); then
        echo "$PAX_ID not in range 0-1"
        exit 1
    fi
}

getPAXTemperature() {
    local SWITCH_NAME=$1

    # Make sure we can get the PAX ID
    if [ ! "$(switchtec temp "$SWITCH_NAME")" ]; then
        echo "Unable to retrieve PAX Temperature"
        exit $?
    fi

    PAX_TEMPERATURE=$(switchtec temp "$SWITCH_NAME")
}

setDeviceName() {
    DRIVES=()
    getDriveList

    for DRIVE in "${DRIVES[@]}";
    do
        SerialNumber=$(nvme id-ctrl "${DRIVE}" | grep -E "^sn " | awk '{print $3}')
        deviceName[$SerialNumber]=$DRIVE
    done
}

displayDriveSlotStatus() {
    local SWITCH_NAME=$1

    getChassis

    # Physical slot ids are set into the hardware. These are the mappings
    declare -a PAX0_DriveSlotFromPhysicalPort=(
        # Drives
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
    declare -a PAX1_DriveSlotFromPhysicalPort=(
        # Drives
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

    getPAXID "$SWITCH_NAME"

    # Associate serial number with its /dev/nvme device
    setDeviceName

    # Grab the attach status of each physical port
    mapfile -t physicalPortStatus < <(switchtec fabric gfms-dump "$SWITCH" | grep " Physical Port ID")

    local physicalPortString
    printf "DEVICE: %s PAX_ID: %d\n\n" "$SWITCH_NAME" "$PAX_ID"
    for physicalPortString in "${physicalPortStatus[@]}";
    do
        local PHYSICAL_PORT_ID
        local MF FW SN Device
        PHYSICAL_PORT_ID=$(echo "$physicalPortString" | awk '{print $4}')
        case $PAX_ID in
            0)
                SLOT=${PAX0_DriveSlotFromPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            1)
                SLOT=${PAX1_DriveSlotFromPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            *)
                exit 1
        esac

        PDFID=$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "$physicalPortString" -A2 | grep "PDFID" | awk '{print $2}')
        if [ -z "$PDFID" ]; then
            PDFID=""
            MF=""
            FW=""
            SN=""
            Device=""
        else
            mapfile -t idCtrl < <(switchtec-nvme id-ctrl "$PDFID"@"$SWITCH_NAME" 2>&1)
            case "${idCtrl[0]}" in
                "NVME Identify Controller:")
                    MF="$(printf '%s\n' "${idCtrl[@]}" | grep -E "^mn " | awk '{print $3}')"
                    FW="$(printf '%s\n' "${idCtrl[@]}" | grep -E "^fr " | awk '{print $3}')"
                    SN="$(printf '%s\n' "${idCtrl[@]}" | grep -E "^sn " | awk '{print $3}')"
                    Device="${deviceName["$SN"]}"
                    ;;
                *)
                    MF="$(echo "${idCtrl[0]}" | awk '{print $3}')"
                    SN=""
                    FW=""
                    Device=""
                    ;;
            esac
        fi

        printf "PDFID: %6.6s\tSLOT: %2.2d  %15.15s %s %s %15.15s %s\n" "${PDFID//}" "${SLOT//}" "$MF" "$SN" "$FW" "$Device" "${physicalPortString//}"
    done
    printf "\n"
}

displayStatus() {
    local SWITCH_NAME=$1

    getChassis

    # Physical slot ids are set into the hardware. These are the mappings
    declare -a PAX0_ConnectedEPToPhysicalPort=(
        # Drives
        [8]="Drive Slot 8               "
        [10]="Drive Slot 7               "
        [12]="Drive Slot 15              "
        [14]="Drive Slot 16              "
        [16]="Drive Slot 17              "
        [18]="Drive Slot 18              "
        [20]="Drive Slot 14              "
        [22]="Drive Slot 13              "
        [48]="Drive Slot 12              "

        # Other Links
        [0]="Interswitch Link           "
        [24]="Rabbit,       ${CHASSIS}r7b0n0"
        [32]="Compute 0,    ${CHASSIS}s0b0n0"
        [34]="Compute 1,    ${CHASSIS}s0b1n0"
        [36]="Compute 2,    ${CHASSIS}s1b0n0"
        [38]="Compute 3,    ${CHASSIS}s1b1n0"
        [40]="Compute 4,    ${CHASSIS}s2b0n0"
        [42]="Compute 5,    ${CHASSIS}s2b1n0"
        [44]="Compute 6,    ${CHASSIS}s3b0n0"
        [46]="Compute 7,    ${CHASSIS}s3b1n0"
    )
    declare -a PAX1_ConnectedEPToPhysicalPort=(
        # Drives
        [8]="Drive Slot 4               "
        [10]="Drive Slot 5               "
        [12]="Drive Slot 6               "
        [14]="Drive Slot 2               "
        [16]="Drive Slot 1               "
        [18]="Drive Slot 9               "
        [20]="Drive Slot 10              "
        [22]="Drive Slot 11              "
        [48]="Drive Slot 3               "

        # Other Links
        [0]="Interswitch Link           "
        [24]="Rabbit,       ${CHASSIS}r7b0n0"
        [32]="Compute 8,    ${CHASSIS}s4b0n0"
        [34]="Compute 9,    ${CHASSIS}s4b1n0"
        [36]="Compute 10,   ${CHASSIS}s5b0n0"
        [38]="Compute 11,   ${CHASSIS}s5b1n0"
        [40]="Compute 12,   ${CHASSIS}s6b0n0"
        [42]="Compute 13,   ${CHASSIS}s6b1n0"
        [44]="Compute 14,   ${CHASSIS}s7b0n0"
        [46]="Compute 15,   ${CHASSIS}s7b1n0"
    )

    getPAXID "$SWITCH_NAME"

    mapfile -t physicalPortIdStrings < <(switchtec status "$SWITCH_NAME" | grep "Phys Port ID:")

    local physicalPortString
    if [ "$VERBOSE" == "true" ]; then
        getPAXTemperature "$SWITCH_NAME"
        printf "DEVICE: %s PAX_ID: %d  TEMP: %s\n\n" "$SWITCH_NAME" "$PAX_ID" "$PAX_TEMPERATURE"
    else
        printf "DEVICE: %s PAX_ID: %d\n\n" "$SWITCH_NAME" "$PAX_ID"
    fi
    printf "Switch Connection        \tStatus\n"
    printf "===========================\t======\n"
    for physicalPortString in "${physicalPortIdStrings[@]}";
    do
        local PHYSICAL_PORT_ID
        PHYSICAL_PORT_ID=$(echo "$physicalPortString" | awk '{print $4}')
        case $PAX_ID in
            0)
                ENDPOINT=${PAX0_ConnectedEPToPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            1)
                ENDPOINT=${PAX1_ConnectedEPToPhysicalPort[$PHYSICAL_PORT_ID]}
                ;;
            *)
                exit 1
        esac

        OPER_STATUS=$(switchtec status "$SWITCH_NAME" | grep "$physicalPortString" -A4 | grep "Status" | awk '{print $2}' )

        printf "%s\t%s\n" "$ENDPOINT" "$OPER_STATUS"
    done
    printf "\n"
}


alias TIME=""
while getopts "tvh:" OPTION
do
    case "${OPTION}" in
        't')
            alias TIME=time
            export TIMEFORMAT='%3lR'
            ;;
        'v')
            export VERBOSE="true"
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

function ep-tunnel-command() {
    local SWITCH=$1 CMD=$2
    echo "Execute switch ep-tunnel $CMD on $SWITCH"

    Endpoints=$(getPDFIDs "$SWITCH")
    for DRIVE in $Endpoints;
    do
        case $CMD in
            status)
                printf "%s\t" "$DRIVE"
                ;;
            *)
                ;;
        esac

        TIME switchtec fabric ep-tunnel-cfg "$SWITCH" --cmd="$CMD" --pdfid="$DRIVE";
    done
}


case $1 in
    slot-info)
        function slot-info() {
            local SWITCH=$1
            echo "Execute slot-info on $SWITCH"
            TIME displayDriveSlotStatus "$SWITCH"
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
            TIME displayStatus "$SWITCH"
        }
        execute status
        ;;
    switchtec-status)
        function switchtec-status() {
            local SWITCH=$1
            echo "Execute switchtec status on $SWITCH"
            TIME switchtec status "$SWITCH"
        }
        execute switchtec-status
        ;;
    ep-tunnel-status)
        function ep-tunnel-status() {
            local SWITCH=$1
            ep-tunnel-command "$SWITCH" "status"
        }
        execute ep-tunnel-status
        ;;
    ep-tunnel-enable)
        function ep-tunnel-enable() {
            local SWITCH=$1
            ep-tunnel-command "$SWITCH" "enable"
        }
        execute ep-tunnel-enable
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
