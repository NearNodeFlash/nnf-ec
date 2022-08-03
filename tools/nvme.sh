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
Usage: $0 COMMAND [ARGS...]

Commands:
    create                               create an nvme namespace on each drive
    attach [NAMESPACE-ID] [CONTROLLER]   attach namespaces from each drive to a controller
    delete [NAMESPACE-ID]                delete an nvme namespace on each drive
EOF
}

SWITCHES=("/dev/switchtec0" "/dev/switchtec1")

case $1 in
    create)
        SIZE=97670000
        for SWITCH in "${SWITCHES[@]}";
        do
            mapfile -t PDFIDS < <(switchtec fabric gfms-dump ${SWITCH} | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}')
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Creating Namespaces on ${PDFIDS[$INDEX]}"
                switchtec-nvme create-ns ${PDFIDS[$INDEX]}@$SWITCH --nsze=$SIZE --ncap=$SIZE --block-size=4096
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
                switchtec-nvme attach-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=$NAMESPACE --controllers=$CONTROLLER
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
                switchtec-nvme delete-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=$NAMESPACE
            done
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac
