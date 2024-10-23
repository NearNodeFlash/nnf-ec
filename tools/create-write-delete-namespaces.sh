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

DURATION=${1:-"30s"}
OP=${2:-"randread"}

usage() {
    cat <<EOF
Create a namespace on every drive, create a logical volume from those namespaces, operate over that logical volume, then delete everything

Usage: $0 [duration: default 30s] [operation: randread|randwrite: default randread]
Examples:
   $0 60s  randread                 # random read test for 60 seconds
   $0 60m  randread                 # random read test for 60 minutes
   $0 300s randwrite                # random write test for 300 seconds
EOF
}

while getopts "h" OPTION
do
    case "${OPTION}" in
        'h')
            usage
            exit 0
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND - 1))

SIZE=0
DELAY_FOR_DEVICES=2

printf "duration %s\n" $DURATION
printf "operation %s\n" $OP

# Ensure lvm.conf doesn't get in the way
sed -i 's/use_lvmlockd = 1/use_lvmlockd = 0/g' /etc/lvm/lvm.conf

# Ensure the /dev/rabbit directory doesn't already exist
rm -rf /dev/rabbit

# Run nnf-ec just to be sure
./nnf-ec

# Create and attach a namespace on each drive to the RABBIT's processor
./nvme.sh create "$SIZE"
./nvme.sh attach

printf "Sleeping %d seconds to allow all devices to be available for logical volume\n" "$DELAY_FOR_DEVICES"
sleep "$DELAY_FOR_DEVICES"

# Create a logical volume spanning all the namespaces
./lvm.sh create

# Write a little something to the logical volume to give the drives some work to do on delete-ns operation
fio --direct=1 --rw="$OP" --bs=32M --ioengine=libaio --iodepth=128 --numjobs=4 --runtime="$DURATION" --time_based --group_reporting --name=rabbit --eta-newline=1 --filename=/dev/rabbit/rabbit

# Delete the logical volume to tidy up
./lvm.sh delete

# Format the namespace to speed up deletion
./nvme.sh cmd format --force --namespace-id=1

# Wait for the format to finish
gbToFormat=$(nvme list | grep KIO | awk '{print $6}' | tr ' ' '\n' | paste -sd+ - | bc | awk '{print ($1 == int($1)) ? int($1) : int($1) + 1}')
while (("$gbToFormat" > 0)); do
    sleep 1
    printf "Formatting, space left %d\n" "$(nvme list | grep KIO | awk '{print $6}' | tr ' ' '\n' | paste -sd+ - | bc | awk '{print ($1 == int($1)) ? $1 : int($1) + 1}')"
    gbToFormat=$(nvme list | grep KIO | awk '{print $6}' | tr ' ' '\n' | paste -sd+ - | bc | awk '{print ($1 == int($1)) ? int($1) : int($1) + 1}')
done

# Show the nvme namespaces for the record
nvme list | grep KIO

# Finally, delete the namespaces
./nvme.sh -t delete
