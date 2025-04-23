#!/bin/bash

# Copyright 2025 Hewlett Packard Enterprise Development LP
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
shopt -s extglob

usage() {
    cat <<EOF
Manage zpool on Rabbit
NOTE: Namespaces must be created and attached prior

Usage: $0 COMMAND [ARGS...]

Commands:
    list-drives                                               list the drives used in create/delete
    create [NAME] [NAMESPACE-ID] [mirror|raidz|raidz2|raidz3] create a zpool on all drives, optionally configuring raidz
    delete [NAME] [NAMESPACE-ID]                              delete a zpool
    status                                                    status of zpool
    scrub                                                     scrub the zpool
    list-volumes                                              list volumes in zpool
EOF
}

DRIVES=()

drives() {
    local NAMESPACE=$1
    DRIVES=()
    echo "Finding drives for namespace ${NAMESPACE}"
    for drive in /dev/disk/by-path/*; do
        if [[ $drive =~ pci-0000:(05|06|07|08|09|0a|0b|0d|83|84|85|86|88|89|8a|8b):00\.0-nvme-"$NAMESPACE"$ ]]; then
            DRIVES+=("$drive")
        fi
    done
    if (( "${#DRIVES[@]}" == 0 )); then
        echo "    No drives found for namespace ${NAMESPACE}"
        return
    fi
    echo "    Found drives for namespace ${NAMESPACE}:"
    for drive in "${DRIVES[@]}"; do
        echo "        $drive"
    done
    echo "    Found ${#DRIVES[@]} drives with namespace ${NAMESPACE}"
}

join() {
    local IFS="$1"
    shift
    echo "$*"
}


NAME=${2:-"rabbit"}
NAMESPACE=${3:-"1"}
RAID_LEVEL=${4:-"raidz2"}

case $1 in
    list-drives)
        drives "$NAMESPACE"
        ;;
    create)
        drives "$NAMESPACE"
        if (( "${#DRIVES[@]}" == 0 )); then
            echo "No drives found, please create and attach namespaces"
            usage
            exit 1
        fi

        zpool create "${NAME}" -o autotrim=on -o cachefile=none  "${RAID_LEVEL}" $(join " " "${DRIVES[@]}")
        if [ $? -ne 0 ]; then
            echo "Failed to create zpool"
            exit 1
        fi
        echo "Created zpool '${NAME}' with ${RAID_LEVEL} on drives: $(join " " "${DRIVES[@]}")"

        echo "DONE! Access the volume at /dev/${NAME}/${NAME}"
        ;;
    delete)
        echo "Removing zpool '${NAME}'"
        zpool destroy "${NAME}" || true
        ;;
    status)
        echo "Status of '${NAME}'"
        zpool status -v -P -L "${NAME}"
        ;;
    scrub)
        echo "Scrubbing '${NAME}'"
        zpool scrub "${NAME}"
        ;;
    list-volumes)
        echo "Volumes in '${NAME}'"
        zpool list -o name,size,alloc,free,cap,dedup,health "${NAME}"
        ;;
    *)
        usage
        exit 1
        ;;
esac
