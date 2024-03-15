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
Simple pass/fail test to determine if Rabbit links are available
Usage: $0

EOF
}

while getopts "hv" OPTION
do
    case "${OPTION}" in
        'h')
            usage
            exit 0
            ;;
        'v')
            verbose="true"
    esac
done
shift $((OPTIND - 1))


drivePhysicalPortConnectedCount=$(lspci -PP -vvv 2>nul | grep -e "^.*\/.*\/.*4200" -A50 | grep -e 4200 -e LnkSta: -e LnkCap: | grep -e "16GT.*Width x4" | wc -l)
drivePhysicalPortDisConnectedCount=$(lspci -PP -vvv 2>nul | grep -e "^.*\/.*\/.*4200" -A50 | grep -e 4200 -e LnkSta: -e LnkCap: | grep -e "Width x0" | wc -l)

if [ "$verbose" == "true" ]
then
    printf "physical port connected count: %d\n" "$drivePhysicalPortConnectedCount"
    printf "physical port disconnected count: %d\n" "$drivePhysicalPortDisConnectedCount"
fi

if [ "$drivePhysicalPortConnectedCount" == "16" ] && [ "$drivePhysicalPortDisConnectedCount" == "2" ]
then
    printf "PASS - %d drives, %d empty slots\n" "$drivePhysicalPortConnectedCount" "$drivePhysicalPortDisConnectedCount"
else
    printf "FAIL - %d drives, %d empty slots\n" "$drivePhysicalPortConnectedCount" "$drivePhysicalPortDisConnectedCount"
fi