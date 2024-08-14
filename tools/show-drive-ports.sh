#!/bin/bash

# Copyright 2024 Hewlett Packard Enterprise Development LP
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
Examine "lspci" status and report failing drive links
Usage: $0 [-h] [-v]

Arguments:
  -v                verbose output
  -h                display this help

Examples:
  $0 -v
EOF
}

VERBOSE="false"
while getopts ":vh" OPTION
do
    case "${OPTION}" in
        'v')
            VERBOSE="true"
            ;;
        'h')
            usage
            exit 0
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND - 1))

drive_links="04 82"

error_present="false"
for pax in $drive_links;
do
    for drive in {0..8};
    do
        if [ $pax -eq 4  ] && [ $drive -eq 7 ]; then
            continue
        fi
        if [ $pax -eq 82  ] && [ $drive -eq 4 ]; then
            continue
        fi

        pci_result=$(lspci -s "$pax":"$drive".0 -vv | grep LnkSta:)
        if [ $? -ne 0 ]; then
            if [ $VERBOSE == "true" ]; then
                printf "%d:%d.0 Error accessing drive\n" "$pax" "$drive";
            fi
            error_present="true"
            continue
        fi

        access_error_result=$(echo "$pci_result" | grep "!!! Unknown header type")
        if [ $? -eq 0 ]; then
            if [ $VERBOSE == "true" ]; then
                printf "%d:%d.0 Unable to retrieve link info\n" "$pax" "$drive"
            fi
            error_present="true"
            continue
        fi

        width_zero_error_result=$(echo "$pci_result" | grep "Width x0")
        if [ $? -eq 0 ]; then
            if [ $VERBOSE == "true" ]; then
                printf "%d:%d.0 Link down\n" "$pax" "$drive"
            fi
            error_present="true"
            continue
        fi

        link_speed_error_result=$(echo "$pci_result" | grep 16GT)
        if [ $? -ne 0 ]; then
            if [ $VERBOSE == "true" ]; then
                printf "%d:%d.0 Link speed too slow\n" "$pax" "$drive"
            fi
            error_present="true"
            continue
        fi
    done;
done

if [ "$error_present" == "false" ]; then
    printf "No errors\n"
else
    printf "ERRORS FOUND\n"
fi
