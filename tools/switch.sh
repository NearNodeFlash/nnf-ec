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
set -eo pipefail
shopt -s expand_aliases

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
    ep-tunnel-disable                    disable the endpoint tunnel for each drive
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

# Retrieve the Physical Device Fabric IDs used to iterate through a list of nvme drives
function getPDFIDs() {
    local SWITCH=$1 FUNCTION="${2:-0}"

    switchtec fabric gfms-dump "$SWITCH" | grep "Function $FUNCTION " -A2 | grep PDFID | awk '{print $2}'
}

function getDriveList() {
    # DRIVES=$1
    # for DRIVE in $(ls /dev/nvme* | grep -E "nvme[[:digit:]]+$");
    for DRIVE in /dev/nvme[0-9]*;
    do
        # shellcheck disable=SC2086
        if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ];
        then
            # SerialNumber=$(nvme id-ctrl ${DRIVE} | grep -E "^sn " | awk '{print $3}')
            # Mfg=$(nvme id-ctrl ${DRIVE} | grep -E "^mn " | awk '{print $3}')
            # FW=$(nvme id-ctrl ${DRIVE} | grep -E "^fr " | awk '{print $3}')
            # printf "%s\t%s\t%s\t%s\n" "$DRIVE" "$Mfg" "$SerialNumber" "$FW"

            DRIVES+=("${DRIVE}")
        fi
    done

    DriveCount="${#DRIVES[@]}"
    if ((DriveCount == 0));
    then
        printf "No drives found: Did you run nnf-ec?\n"
    fi
}

getChassis() {
    if [ "$VERBOSE" != "true" ]; then
        CHASSIS="       "
        return
    fi

    COMMAND=xhost-query.py
    if command -v $COMMAND &> /dev/null; then
        CHASSIS=$("$COMMAND" "$(hostname)" | cut -c -7)
    else
        CHASSIS=$(hostname | cut -c -8)
    fi
}

getPAXID() {
    local SWITCH_NAME=$1

    PAX_ID=$(switchtec fabric gfms-dump "$SWITCH_NAME" | grep "^PAX ID:" | awk '{print $3}')
    ret=$?
    if [ ! $ret ]; then
        echo "Unable to retrieve PAX ID"
        exit $ret
    fi

    if ! (( PAX_ID >= 0 && PAX_ID <= 1 )); then
        echo "$PAX_ID not in range 0-1"
        exit 1
    fi
}

getPAXTemperature() {
    local SWITCH_NAME=$1

    PAX_TEMPERATURE=$(switchtec temp "$SWITCH_NAME")
    ret=$?
    if [ ! $ret ]; then
        echo "Unable to retrieve PAX Temperature"
        exit $ret
    fi
}

displayPAX() {
    local SWITCH_NAME=$1
    getPAXID "$SWITCH_NAME"

    if [ "$VERBOSE" == "true" ]; then
        getPAXTemperature "$SWITCH_NAME"
        printf "DEVICE: %s PAX_ID: %d  TEMP: %s\n\n" "$SWITCH_NAME" "$PAX_ID" "$PAX_TEMPERATURE"
    else
        printf "DEVICE: %s PAX_ID: %d\n\n" "$SWITCH_NAME" "$PAX_ID"
    fi
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
        # [22]=13         SLOT 13 is not supported
        [48]=12
    )
    declare -a PAX1_DriveSlotFromPhysicalPort=(
        # Drives
        [8]=4
        [10]=5
        [12]=6
        [14]=2
        # [16]=1          SLOT 1 is not supported
        [18]=9
        [20]=10
        [22]=11
        [48]=3
    )

    # Associate serial number with its /dev/nvme device
    setDeviceName

    # Grab the attach status of each physical port
    mapfile -t physicalPortStatus < <(switchtec fabric gfms-dump "$SWITCH" | grep " Physical Port ID")

    local physicalPortString
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

        if [ -n "$SLOT" ]
        then
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
        fi
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
        # [22]="Drive Slot 13              "        SLOT 13 is not supported
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
        # [16]="Drive Slot 1               "        SLOT 1 is not supported
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

    # Example switchtec status in table format
    #
    #       [root@x9000c3j7b0n0 tools]# ./switch.sh cmd status --format=table
    #       DEVICE: /dev/switchtec0 PAX_ID: 1
    #
    #       [00] part:00.01 w:cfg[x16]-neg[x16] stk:0.0 lanes:0123456789abcdef rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [08] part:00.02 w:cfg[x04]-neg[x04] stk:1.0 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [10] part:00.03 w:cfg[x04]-neg[x04] stk:1.2 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [12] part:00.04 w:cfg[x04]-neg[x04] stk:1.4 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [14] part:00.05 w:cfg[x04]-neg[x04] stk:1.6 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [16] part:00.06 w:cfg[x04]-neg[x00] stk:2.0 lanes:xxxx             rev:0 dsp link:0 rate:G1 LTSSM:Detect (QUIET)
    #       [18] part:00.07 w:cfg[x04]-neg[x04] stk:2.2 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [20] part:00.08 w:cfg[x04]-neg[x04] stk:2.4 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [22] part:00.09 w:cfg[x04]-neg[x04] stk:2.6 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [24] part:01.00 w:cfg[x16]-neg[x16] stk:3.0 lanes:fedcba9876543210 rev:1 usp link:1 rate:G4 LTSSM:L0 (L0)
    #       [32] part:02.00 w:cfg[x04]-neg[x00] stk:4.0 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [34] part:03.00 w:cfg[x04]-neg[x00] stk:4.2 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [36] part:04.00 w:cfg[x04]-neg[x00] stk:4.4 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [38] part:05.00 w:cfg[x04]-neg[x00] stk:4.6 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [40] part:06.00 w:cfg[x04]-neg[x00] stk:5.0 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Detect (QUIET)
    #       [42] part:07.00 w:cfg[x04]-neg[x00] stk:5.2 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [44] part:08.00 w:cfg[x04]-neg[x00] stk:5.4 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [46] part:09.00 w:cfg[x04]-neg[x00] stk:5.6 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [48] part:00.10 w:cfg[x04]-neg[x04] stk:6.0 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       DEVICE: /dev/switchtec1 PAX_ID: 0
    #
    #       [00] part:00.01 w:cfg[x16]-neg[x16] stk:0.0 lanes:fedcba9876543210 rev:1 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [08] part:00.02 w:cfg[x04]-neg[x04] stk:1.0 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [10] part:00.03 w:cfg[x04]-neg[x04] stk:1.2 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [12] part:00.04 w:cfg[x04]-neg[x04] stk:1.4 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [14] part:00.05 w:cfg[x04]-neg[x04] stk:1.6 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [16] part:00.06 w:cfg[x04]-neg[x04] stk:2.0 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [18] part:00.07 w:cfg[x04]-neg[x04] stk:2.2 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [20] part:00.08 w:cfg[x04]-neg[x04] stk:2.4 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)
    #       [22] part:00.09 w:cfg[x04]-neg[x00] stk:2.6 lanes:xxxx             rev:0 dsp link:0 rate:G1 LTSSM:Detect (QUIET)
    #       [24] part:01.00 w:cfg[x16]-neg[x16] stk:3.0 lanes:fedcba9876543210 rev:1 usp link:1 rate:G4 LTSSM:L0 (L0)
    #       [32] part:02.00 w:cfg[x04]-neg[x00] stk:4.0 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [34] part:03.00 w:cfg[x04]-neg[x00] stk:4.2 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [36] part:04.00 w:cfg[x04]-neg[x00] stk:4.4 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [38] part:05.00 w:cfg[x04]-neg[x00] stk:4.6 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [40] part:06.00 w:cfg[x04]-neg[x00] stk:5.0 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [42] part:07.00 w:cfg[x04]-neg[x00] stk:5.2 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [44] part:08.00 w:cfg[x04]-neg[x00] stk:5.4 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [46] part:09.00 w:cfg[x04]-neg[x00] stk:5.6 lanes:xxxx             rev:0 usp link:0 rate:G1 LTSSM:Polling (COMP)
    #       [48] part:00.10 w:cfg[x04]-neg[x04] stk:6.0 lanes:0123             rev:0 dsp link:1 rate:G4 LTSSM:L0 (L0)

    mapfile -t statusTableLines < <(switchtec status --format=table "$SWITCH_NAME")

    printf "Switch Connection        \tStatus\tRate\tWidth\n"
    printf "===========================\t======\t====\t=================\n"

    # Show the downstream ports (drives) before the upstream (computes and Rabbit-p)
    local PORT_DIRECTION=("dsp" "usp")
    for PD in "${PORT_DIRECTION[@]}";
    do
        for line in "${statusTableLines[@]}";
        do
            linePortDirection=$(echo "$line" | awk '{print $7}')
            if [ "$PD" = "$linePortDirection" ]
            then
                PHYSICAL_PORT_ID=$(echo "$line" | awk '{gsub(/\[|\]/, "", $1); num=sprintf("%d", $1); print num}')
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

                if [ -n "$ENDPOINT" ]
                then
                    # Add an asterix "*" for PCI WIDTHs or RATEs not matching the expected values
                    WIDTH=$(echo "$line" | awk '{gsub(/w:/, "", $3); if (match($3, /cfg\[x([0-9]+)\]-neg\[x([0-9]+)\]/, arr)) {
                                var1 = arr[1];
                                var2 = arr[2];
                            }
                            if (var1 == var2) print $3; else print $3"*";}')
                    LINK=$(echo "$line" | awk '{gsub(/link:/, "", $8); if ($8 == 1) print "UP"; else if ($8 == 0) print "DOWN"; }')
                    RATE=$(echo "$line" | awk '{gsub(/rate:/, "", $9); if ($9 == "G4") print $9; else print $9"*"; }')
                    printf "%s\t%s\t%s\t%s\n" "$ENDPOINT" "$LINK" "$RATE" "$WIDTH"
                fi
            fi
        done
        printf "\n"
    done
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
            displayPAX "$SWITCH"
            TIME displayDriveSlotStatus "$SWITCH"
        }
        execute slot-info
        ;;
    info)
        function info() {
            local SWITCH=$1
            displayPAX "$SWITCH"
            TIME switchtec info "$SWITCH"
        }
        execute info
        ;;
    status)
        function status() {
            local SWITCH=$1
            # echo "Execute switch status on $SWITCH"
            displayPAX "$SWITCH"
            TIME displayStatus "$SWITCH"
        }
        execute status
        ;;
    switchtec-status)
        function switchtec-status() {
            local SWITCH=$1
            # echo "Execute switchtec status on $SWITCH"
            displayPAX "$SWITCH"
            TIME switchtec status "$SWITCH"
        }
        execute switchtec-status
        ;;
    ep-tunnel-status)
        function ep-tunnel-status() {
            local SWITCH=$1
            displayPAX "$SWITCH"
            ep-tunnel-command "$SWITCH" "status"
        }
        execute ep-tunnel-status
        ;;
    ep-tunnel-enable)
        function ep-tunnel-enable() {
            local SWITCH=$1
            displayPAX "$SWITCH"
            ep-tunnel-command "$SWITCH" "enable"
        }
        execute ep-tunnel-enable
        ;;
    ep-tunnel-disable)
        function ep-tunnel-disable() {
            local SWITCH=$1
            displayPAX "$SWITCH"
            ep-tunnel-command "$SWITCH" "disable"
        }
        execute ep-tunnel-disable
        ;;
    fabric)
        function fabric() {
            local SWITCH=$1 FABRIC_CMD=$2 ARGS=( "${@:3}" )
            if [ "$VERBOSE" == "true" ]; then echo "Execute switch fabric $FABRIC_CMD"; fi
            displayPAX "$SWITCH"
            TIME switchtec fabric "$FABRIC_CMD" "$SWITCH" "${ARGS[@]}"
        }
        execute fabric "${2:-gfms-dump}" "${@:3}"
        ;;
    cmd)
        function cmd() {
            local SWITCH=$1 CMD=$2 ARGS=( "${@:3}" )
            displayPAX "$SWITCH"
            TIME switchtec "$CMD" "$SWITCH" "${ARGS[@]}"
        }
        execute cmd "${@:2}"
        ;;
    *)
        usage
        exit 1
        ;;
esac
