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
set -eo pipefail
shopt -s expand_aliases
shopt -s extglob

usage() {
    cat <<EOF
Manage LVM volumes on Rabbit
NOTE: Namespaces must be created and attached prior

Usage: $0 COMMAND [ARGS...]

Commands:
    list-drives                                         list the drives used in create/delete
    create [NAME] [NAMESPACE-ID] [striped|raid5|raid6]  create an LVM volume all drives, optionally configuring RAID
    delete [NAME] [NAMESPACE-ID]                        delete an LVM volume
    status                                              status of LVM volumes
EOF
}

DRIVES=()

drives() {
    local NAMESPACE=$1
    DRIVES=()
    for DRIVE in /dev/nvme*n"$NAMESPACE";
    do
        # shellcheck disable=SC2086
        if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ];
        then
            echo "  Found drive ${DRIVE}"
            NAMESPACEID=$(nvme id-ns ${DRIVE} | grep -E '^NVME Identify Namespace [[:digit:]]+' | awk '{print $4}')
            if [ "${NAMESPACEID::-1}" == "$NAMESPACE" ];
            then
                echo "    Found Namespace ${NAMESPACE}"
                DRIVES+=("${DRIVE}")
            fi
        fi
    done

    echo "${#DRIVES[@]}" DRIVES: "${DRIVES[@]}"
}

join() {
    local IFS="$1"
    shift
    echo "$*"
}


NAME=${2:-"rabbit"}
NAMESPACE=${3:-"1"}
RAID_LEVEL=${4:-"striped"}

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

        for DRIVE in "${DRIVES[@]}";
        do
            echo "Creating Physical Volume '${DRIVE}'"
            pvcreate "${DRIVE}"
        done

        echo "Creating Volume Group '${NAME}'"
        # shellcheck disable=2046
        vgcreate "${NAME}" $(join " " "${DRIVES[@]}")

        echo "Activating Volume Group '${NAME}'"
        vgchange --activate y "${NAME}"

        TOTAL_STRIPES=$(( ${#DRIVES[@]} ))
        case "$RAID_LEVEL" in
            raid5)
                PARITY_STRIPES=1
                ;;
            raid6)
                PARITY_STRIPES=2
                ;;
            raid10)
                PARITY_STRIPES=$(( "$TOTAL_STRIPES"/2 ))
                ;;
            *)
                PARITY_STRIPES=0
                ;;
        esac
        STRIPES=$(( "$TOTAL_STRIPES" - "$PARITY_STRIPES" ))

        echo "Creating '$RAID_LEVEL' Logical Volume '${NAME}' with '${STRIPES}' stripes"
        lvcreate --zero n --activate n --extents 100%VG -i "$STRIPES" --stripesize 32KiB --type "$RAID_LEVEL" --name "${NAME}" --noudevsync "${NAME}"

        echo "Activate Volume Group '${NAME}'"
        vgchange --activate y "${NAME}"

        echo "Status '${NAME}'"
        lvs -a -o +devices,raid_sync_action

        echo "DONE! Access the volume at /dev/${NAME}/${NAME}"
        ;;
    delete)
        echo "Removing Logical Volume '${NAME}'"
        lvremove --yes /dev/"${NAME}"/"${NAME}" || true

        echo "Deactivate Volume Group '${NAME}'"
        vgchange --activate n "${NAME}" || true

        echo "Removing Volume Group'${NAME}'"
        vgremove --yes "${NAME}" || true

        drives "$NAMESPACE"
        for DRIVE in "${DRIVES[@]}";
        do
            echo "Remove Physical Volume '${DRIVE}'"
            pvremove "${DRIVE}" || true
        done
        ;;
    status)
        echo "Status of '${NAME}'"
        lvs -a -o +devices,raid_sync_action
        ;;
    *)
        usage
        exit 1
        ;;
esac

