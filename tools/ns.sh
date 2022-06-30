#!/bin/bash
#
# Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
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
Example NVMe Namespace commands running against switchtec managed devices.
Usage: $0 COMMAND

Commands:
    create-ns           create two namespaces on all drives
    attach-ns           attach two namespaces to the controller
    delete-ns           delete two namespaces on all drives
EOF
}

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
for SWITCH in ${SWITCHES[@]}
do
    if [ ! -f $SWITCH ]; then
        echo "Switch $SWITCH does not exists. Are you running from the Rabbit? Is the switch up?"
        exit 1
    fi
done

case $1 in
    create-ns)
        SIZE=97670000
        for SWITCH in ${SWITCHES[@]};
        do
            PDFIDS=( $(switchtec fabric gfms-dump $SWITCH | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}') )
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Creating Namespaces on ${PDFIDS[$INDEX]}"
                switchtec-nvme create-ns ${PDFIDS[$INDEX]}@$SWITCH --nsze=$SIZE --ncap=$SIZE --block-size=4096
                switchtec-nvme create-ns ${PDFIDS[$INDEX]}@$SWITCH --nsze=$SIZE --ncap=$SIZE --block-size=4096
            done
        done
        ;;
    attach-ns)
        SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
        for SWITCH in ${SWITCHES[@]};
        do
            PDFIDS=( $(switchtec fabric gfms-dump $SWITCH | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}') )
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Attaching Namespaces on ${PDFIDS[$INDEX]}"
                switchtec-nvme attach-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=1 --controllers=1
                switchtec-nvme attach-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=2 --controllers=1
            done
        done
        ;;
    delete-ns)
        for SWITCH in ${SWITCHES[@]};
        do
            PDFIDS=( $(switchtec fabric gfms-dump $SWITCH | grep "Function 0 " -A1 | grep PDFID | awk '{print $2}') )
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Deleting Namespaces on ${PDFIDS[$INDEX]}"
                switchtec-nvme delete-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=1
                switchtec-nvme delete-ns ${PDFIDS[$INDEX]}@$SWITCH --namespace-id=2
            done
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac