#!/bin/bash

# Copyright 2023 Hewlett Packard Enterprise Development LP
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
set -e
# set -o xtrace
shopt -s expand_aliases

usage() {
    cat <<EOF
Load tools and firmware files onto Rabbits.
This script assumes it is part of a package of firmware, tool and script directories.

Usage: $0 [-h]

Arguments:
  -h                display this help
EOF
}

while getopts "h:" OPTION
do
    case "${OPTION}" in
        'h',*)
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND - 1))

TLD="/root/ajf"

# Find the list of Rabbits we need to tool-up
rabbits=( $(cm node show | grep Rabbit) )
# rabbits=( RabbitP-c0 )
for rabbit in "${rabbits[@]}";
do
    ssh "$rabbit" "rm -rf tools scripts"
    printf "$rabbit\n" "$rabbit"
    rsync -r "$TLD"/KIOXIA "$rabbit":
    rsync -r "$TLD"/Switchtec "$rabbit":
    rsync "$TLD"/nnf-ec "$rabbit":
    ssh "$rabbit" "chmod +x nnf-ec"
    rsync -r "$TLD"/scripts "$rabbit":
    rsync -r "$TLD"/tools "$rabbit":
    rsync -r "$TLD"/switchtec* "$rabbit":/usr/src

    # Compile/install switchtec-user, then copy the latest switchtec-nvme program into place to get the latest capabilities
    ssh "$rabbit" "cd /usr/src/switchtec-user && ./configure && make install && cp /usr/src/switchtec-nvme-cli/switchtec-nvme /usr/sbin"

    # Launch nnf-ec to initialize PAX's, drives and compute node endpoints
    # type CTRL-C when you see "Starting HTTP Server	{"address": ":8080"}"
    # to proceed to the next Rabbit.
    ssh "$rabbit" "./nnf-ec"
done
