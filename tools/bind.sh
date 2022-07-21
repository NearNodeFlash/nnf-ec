#!/bin/bash

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
Run bind or unbind operations
Usage: $0 [-p] COMMAND

Commands:
    bind            bind all drives to the rabbit
    unbind          unbind all drives from the rabbit

Options:
    -p              Bind physical function (EXPERIMENTAL)
EOF
}


FUNCTION=1
while getopts "ph" OPTION
do
    case "${OPTION}" in
        'p')
            echo "WARNING: Binding of Physical Function is not advised"
            FUNCTION=0
            ;;
        'h')
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND - 1))

SWITCHES=("/dev/switchtec0" "/dev/switchtec1")

case $1 in
    bind)
        for SWITCH in ${SWITCHES[@]};
        do
            HOST_SW_INDEX=$(switchtec fabric gfms-dump $SWITCH | head -n1 | grep "PAX ID" | awk '{print $3}')
            PDFIDS=( $(switchtec fabric gfms-dump $SWITCH | grep "Function $FUNCTION " -A1 | grep PDFID | awk '{print $2}') )
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Performing Bind Operation $SWITCH $HOST_SW_INDEX $INDEX ${PDFIDS[$INDEX]}"
                switchtec fabric gfms-bind $SWITCH --host_sw_idx=$HOST_SW_INDEX --phys_port_id=24 --log_port_id=$INDEX --pdfid=${PDFIDS[$INDEX]}
            done
        done
        ;;
    unbind)
        for SWITCH in ${SWITCHES[@]};
        do
            HOST_SW_INDEX=$(switchtec fabric gfms-dump $SWITCH | head -n1 | grep "PAX ID" | awk '{print $3}')
            PDFIDS=( $(switchtec fabric gfms-dump $SWITCH | grep "Function $FUNCTION " -A1 | grep PDFID | awk '{print $2}') )
            for INDEX in "${!PDFIDS[@]}";
            do
                echo "Performing Unbind Operation $SWITCH $HOST_SW_INDEX $INDEX ${PDFIDS[$INDEX]}"
                switchtec fabric gfms-unbind $SWITCH --host_sw_idx=$HOST_SW_INDEX --phys_port_id=24 --log_port_id=$INDEX
            done
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac
