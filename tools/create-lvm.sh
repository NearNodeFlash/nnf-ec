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
Create a namespace on every drive, create a logical volume from those namespaces
Usage: $0 [NAMESPACE-SIZE-IN-BYTES]
EOF
}

SIZE=${1:-100000000000}
DELAY_FOR_DEVICES=2
DEVICE_COUNT=${2:-1}

case "$SIZE" in
    ''|*[!0-9]*)
        usage
        exit 1
    ;;
    *)
        echo "Device count" "$DEVICE_COUNT"
        for (( dev=1; dev <= DEVICE_COUNT; dev++ ));
        do

            # Create and attach a namespace on each drive to the RABBIT's processor
            ./nvme.sh create "$SIZE"
            ./nvme.sh attach "$dev"

            printf "Sleeping %d seconds to allow all devices to be available for logical volume\n" "$DELAY_FOR_DEVICES"
            sleep "$DELAY_FOR_DEVICES"

            # Create a logical volume spanning all the namespaces
            ./lvm.sh create lvm"$dev" "$dev"
        done
    ;;
esac
