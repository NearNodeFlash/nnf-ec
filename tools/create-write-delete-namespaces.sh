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
Create a namespace on every drive, create a logical volume from those namespaces, write the logical volume, then delete everything
Usage: $0 [NAMESPACE-SIZE-IN-BYTES]

NAMESPACE-SIZE-IN-BYTES: 0 -> (Allocate entire drive)
EOF
}

SIZE=${1:-0}
DELAY_FOR_DEVICES=2
DELAY_TO_COOL_CPU=0

case "$SIZE" in
    ''|*[!0-9]*)
        usage
        exit 1
    ;;
    *)
        # Create and attach a namespace on each drive to the RABBIT's processor
        ./nvme.sh create "$SIZE"
        ./nvme.sh attach

        printf "Sleeping %d seconds to allow all devices to be available for logical volume\n" "$DELAY_FOR_DEVICES"
	    sleep "$DELAY_FOR_DEVICES"

        # Create a logical volume spanning all the namespaces
        ./lvm.sh create

        # Write a little something to the logical volume to give the drives some work to do on delete-ns operation
        fio --direct=1 --rw=randwrite --bs=32M --ioengine=libaio --iodepth=128 --numjobs=4 --runtime=30s --time_based --group_reporting --name=rabbit --eta-newline=1 --filename=/dev/rabbit/rabbit

        if [ $DELAY_TO_COOL_CPU != 0 ]
        then
            printf "Sleeping %d seconds to cool the CPU\n" "$DELAY_TO_COOL_CPU"
            sleep "$DELAY_TO_COOL_CPU"
        fi
        # Delete the logical volume to tidy up
        ./lvm.sh delete

        # Show the nvme namespaces for the record
        nvme list | grep KIO

        # Finally, delete the namespaces
        ./nvme.sh -t delete
    ;;
esac
