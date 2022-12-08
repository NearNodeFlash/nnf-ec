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
Run various NVMe Namespace commands over the NVMe switch fabric using
switchtec-nvme utility.

Usage: $0 [-h] [-t] COMMAND [ARGS...]

Commands:
    create [SIZE-IN-BYTES]               create an nvme namespace on each drive of the specified size. (0 implies max capacity)
    attach [NAMESPACE-ID] [CONTROLLER]   attach namespaces from each drive to a controller
    delete [NAMESPACE-ID]                delete an nvme namespace on each drive
    detach [NAMESPACE-ID] [CONTROLLER]   detach namespaces from each drive for a controller
    list                                 list the nvme namespaces on each drive
    list-pdfid                           list the physical device fabric ID (PDFID) of each drive

    cmd [COMMAND] [ARG [ARG]...]         execute COMMAND on each drive in the fabric.
                                         i.e. $0 id-ctrl

Arguments:
  -h                display this help
  -t                time each command

Examples:
  ./nvme.sh -t delete 1                                                 # delete namespace 1

  ./nvme.sh cmd id-ctrl | grep -E "^fr "                                # display firmware level
  ./nvme.sh cmd id-ctrl | grep -E "^mn "                                # display model name
  ./nvme.sh cmd id-ctrl | grep -e "Execute" -e "^fr " -e "^sn "         # display the drive's PDFID, firmware version, and serial number

  ./nvme.sh cmd format --force --ses=0 --namespace-id=<namespace id>    # format specified namespace
  ./nvme.sh cmd list-ctrl --namespace-id=<ns-id>                        # list the controller attached to namespace "ns-id"

  ./nvme.sh cmd virt-mgmt --cntlid=3 --act=9                            # enable virtual functions for Rabbit

Drive Firmware upgrade:
  ./nvme.sh cmd fw-download --fw=<filename>.ftd                         # download firmware
  ./nvme.sh cmd fw-activate --action=3                                  # activate latest firmware download

EOF
}

# execute <fn<path>> <args...>
execute() {
    local FUNCTION=$1 ARGS=( "${@:2}" )

    if [ "$(type -t $FUNCTION)" != "function" ]; then
        echo "$1 is not a function."
        exit 1
    fi

    SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
    for SWITCH in "${SWITCHES[@]}";
    do
        mapfile -t PDFIDS < <(switchtec fabric gfms-dump "${SWITCH}" | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
        for INDEX in "${!PDFIDS[@]}";
        do
            "$FUNCTION" "${PDFIDS[$INDEX]}@$SWITCH" "${ARGS[@]}"
        done
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
    create)
        function create_ns() {
            local DRIVE=$1 SIZE=$2

            if [ "$SIZE" == "0" ]; then
                echo "Reading Capacity on $DRIVE"
                SIZE=$(switchtec-nvme id-ctrl "$DRIVE" | grep tnvmcap | awk '{print $3}')
            fi

            declare -i SECTORS=$SIZE/4096
            echo "Creating Namespaces on $DRIVE with size ${SIZE}"
            TIME switchtec-nvme create-ns "$DRIVE" --nsze="$SECTORS" --ncap="$SECTORS" --block-size=4096 --nmic=1
        }
        execute create_ns "${2:-0}"
        ;;
    attach)
        function attach_ns() {
            local DRIVE=$1 NAMESPACE=$2 CONTROLLER=$3
            echo "Attaching Namespace $NAMESPACE on $DRIVE to Controller $CONTROLLER"
            TIME switchtec-nvme attach-ns "$DRIVE" --namespace-id="$NAMESPACE" --controllers="$CONTROLLER"
        }
        execute attach_ns "${2:-1}" "${3:-3}"
        ;;
    delete)
        function delete_ns() {
            local DRIVE=$1 NAMESPACE=$2
            echo "Deleting Namespaces $NAMESPACE on $DRIVE"
            TIME switchtec-nvme delete-ns "$DRIVE" --namespace-id="$NAMESPACE"
        }
        execute delete_ns "${2:-1}"
        ;;
    detach)
        function detach_ns() {
            local DRIVE=$1 NAMESPACE=$2 CONTROLLER=$3
            echo "Detaching Namespace $NAMESPACE on $DRIVE from Controller $CONTROLLER"
            TIME switchtec-nvme detach-ns "$DRIVE" --namespace-id="$NAMESPACE" --controllers="$CONTROLLER"
        }
        execute detach_ns "${2:-1}" "${3:-3}"
        ;;
    list)
        function list_ns() {
            local DRIVE=$1
            echo "Namespaces on $DRIVE"
            TIME switchtec-nvme list-ns "$DRIVE" --all
        }
        execute list_ns
        ;;
    list-pdfid)
        function list_pfid() {
            local DRIVE=$1
            echo "$DRIVE"
        }
        execute list_pfid
        ;;

    cmd)
        function cmd() {
            local DRIVE=$1 CMD=$2 ARGS=( "${@:3}" )
            echo "Execute on $DRIVE $CMD" "${ARGS[@]}"
            TIME switchtec-nvme "$CMD" "$DRIVE" "${ARGS[@]}"
        }
        execute cmd "${@:2}"
        ;;
    *)
        usage
        exit 1
        ;;
esac
