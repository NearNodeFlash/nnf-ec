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

usage() {
    cat <<EOF
Run various NVMe Namespace commands
Usage: $0 [-h] [-t] COMMAND [ARGS...]

Commands:
    create [SIZE-IN-BYTES]               create an nvme namespace on each drive of the specified size. (0 implies max capacity)
    attach [NAMESPACE-ID] [CONTROLLER]   attach namespaces from each drive to a controller
    delete [NAMESPACE-ID]                delete an nvme namespace on each drive
    list                                 display all nvme namespaces on each drive

Arguments:
  -h                display this help
  -t                time each command

Example:
  nvme.sh -t delete 1
EOF
}

shopt -s expand_aliases
export TIMEFORMAT='%3lR'
SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
alias TIME=""

while getopts "th:" OPTION
do
    case "${OPTION}" in
        't')
            alias TIME=time
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
        SIZE=${2:-0}
        for SWITCH in "${SWITCHES[@]}";
        do
            mapfile -t PDFIDS < <(switchtec fabric gfms-dump ${SWITCH} | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
            for INDEX in "${!PDFIDS[@]}";
            do
                if [ "$SIZE" == "0" ]; then
                    SIZE=$(switchtec-nvme id-ctrl ${PDFIDS[$INDEX]}@$SWITCH | grep tnvmcap | awk '{print $3}')
                fi

                declare -i SECTORS=$SIZE/4096
                echo "Creating Namespaces on ${PDFIDS[$INDEX]} with size ${SIZE}"
                TIME switchtec-nvme create-ns ${PDFIDS[$INDEX]}@$SWITCH --nsze=$SECTORS --ncap=$SECTORS --block-size=4096
            done
        done
        ;;
    attach)
        NAMESPACE=${2:-"1"}
        CONTROLLER=${3:-"3"}
        for SWITCH in "${SWITCHES[@]}";
        do
            mapfile -t PDFIDS < <(switchtec fabric gfms-dump ${SWITCH} | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Attaching Namespace $NAMESPACE on ${PDFIDS[$INDEX]} to Controller $CONTROLLER"
                TIME switchtec-nvme attach-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=$NAMESPACE --controllers=$CONTROLLER
            done
        done
        ;;
    delete)
        NAMESPACE=${2:-"1"}
        for SWITCH in "${SWITCHES[@]}";
        do
            mapfile -t PDFIDS < <(switchtec fabric gfms-dump ${SWITCH} | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Deleting Namespaces $NAMESPACE on ${PDFIDS[$INDEX]}"
                TIME switchtec-nvme delete-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=$NAMESPACE
            done
        done
        ;;
    list)
        for SWITCH in "${SWITCHES[@]}";
        do
            mapfile -t PDFIDS < <(switchtec fabric gfms-dump ${SWITCH} | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Namespaces on ${PDFIDS[$INDEX]}"
                TIME switchtec-nvme list-ns ${PDFIDS[$INDEX]}@$SWITCH --all
            done
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac
