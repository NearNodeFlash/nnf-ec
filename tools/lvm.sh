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
Manage LVM volumes on Rabbit
Usage: $0 COMMAND [ARGS...]

Commands:
    list-drives                         list the drives used in create/delete
    create [NAME] [NAMESPACE-ID]        create an LVM volume all drives
    delete [NAME]                       delete an LVM volume
EOF
}

DRIVES=()

drives() {
    local NAMESPACE=$1
    for DRIVE in $(ls /dev/nvme* | grep -E "nvme[[:digit:]]+n[[:digit:]]+$");
    do
        if [ "$(nvme id-ctrl ${DRIVE} | grep KIOXIA)" != "" ];
        then
            echo "  Found Kioxia drive ${DRIVE}"
            NAMESPACEID=$(nvme id-ns ${DRIVE} | grep -E '^NVME Identify Namespace [[:digit:]]+' | awk '{print $4}')
            if [ "${NAMESPACEID::-1}" == "$NAMESPACE" ];
            then
                echo "    Found Namespace ${NAMESPACE}"
            
                DRIVES+="${DRIVE} "
            fi
        fi
    done

    echo "DRIVES: ${DRIVES[@]}"
}

NAME=${2:-"rabbit"}
NAMESPACE=${3:-"1"}

case $1 in
    list-drives)
        drives $NAMESPACE
        ;;
    create)
        drives $NAMESPACE
        for DRIVE in ${DRIVES[@]};
        do
            echo "Creating Physical Volume '${DRIVE}'"
            pvcreate ${DRIVE}
        done

        echo "Creating Volume Group '${NAME}'"
        vgcreate ${NAME} ${DRIVES[@]}

        echo "Creating Logical Volume '${NAME}'"
        lvcreate -Zn --extents 100%VG --stripes ${#DRIVES[@]} --stripesize 32KiB --name ${NAME} ${NAME}

        echo "Activate Volume Group '${NAME}'"
        vgchange --activate y ${NAME}

        echo "DONE! Access the volume at /dev/${NAME}/${NAME}"
        ;;
    delete)
        echo "Removing Logical Volume '${NAME}'"
        lvremove --yes /dev/{$NAME}/${NAME}

        echo "Deactivate Volume Group '${NAME}'"
        vgchange --activate n ${NAME}

        echo "Removing Volume Group'${NAME}'"
        vgremove --yes ${NAME}

        drives $NAMESPACE
        for DRIVE in ${DRIVES};
        do
            echo "Remove Physical Volume '${DRIVE}'"
            pvremove --yes ${DRIVE}
        done
        ;;
    *)
        usage
        exit 1
        ;;
esac

