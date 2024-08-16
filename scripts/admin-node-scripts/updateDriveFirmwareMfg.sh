#!/bin/bash

# Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
#=================================================================
# The following assumtions are made/expected coming into this script:
#   * The Rabbits are installed in the Cabinet with geo-location "x1002c?j4b0", "x1002c?j7b0", and "x1002c?j7b0n0"
#   * The Rabbits are in the HPCM database & have had their SSH keys set
#   * The Rabbits nC's are fully booted
#
# ============ This is Ver1.6 of the script, 8/15/2024 ===========
#  This script expects and/or programs the following FW versions:
#    E3.s Kioxia Drive FW = ver 1TCRS104
#===================================================================
set -e
# set -o xtrace

usage() {                                                                          # This function is called if a bad/missing parameter is found - displays proper usage
    cat <<EOF                                                                      # The 'EOF' is a "Here Tag" - will 'cat' all the text until an EOF is found
Query drive firmware version for all drives in a Rabbit. Update drives that are out-of-date.

Assumes that the following are installed on the Rabbit:
- /root/nnf-ec
- /root/tools/nvme.sh
- <firmware-file-path>

Usage: $0 [-h] [RABBIT-XNAME] [EXPECTED-FIRMWARE] [FIRMWARE-FILE-PATH]

Arguments:
  -h                display this help

Examples:
    ./updateDriveFirmware.sh -h                                                     # Display help message
    ./updateDriveFirmware.sh x1002c3j7b0n0 1TCRS104 /root/KIOXIA/1TCRS104.std       # Rabbit: x1002c3j7b0n0, Expected Firmware: "1TCRS104", Firmware File Path: "x1002c3rbt7b0n0:/root/KIOXIA/1TCRS104.std"
EOF
}

alias TIME=""
while getopts "th" OPTION;
do
    case "${OPTION}" in
        t)
            alias TIME=time
            export TIMEFORMAT='%3lR'
            ;;
        h)
            usage
            exit 0
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND - 1))

if [ $# -lt 3 ]; then
    usage
    exit 1
fi

rabbit=$1
expectedFirmware=$2
firmwareFile=$3
LOGFILE="$(pwd)/logs/$rabbit.log"
LOGFILE_FAILURE="$(pwd)/logs/${rabbit}_Failure.log"
TEE_LOGFILE="tee -a $LOGFILE"

echo -e "     Validating firmware $expectedFirmware is on Rabbit's drives . . ." > "$LOGFILE"

# Run nnf-ec to initialize PAX chips and drives
echo -e "     Initialize PCIe Switch connections to drives first:"  | eval "$TEE_LOGFILE"

if [ ! "$(ssh "$rabbit" ./nnf-ec -initializeAndExit 2>&1 >/dev/null)" ]
then
     DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                            # Date/time stamp for the log file
     echo -e "\nBye-bye with an NNF-EC Failure at $DATE_TIME!\n"  | eval "$TEE_LOGFILE"
     cp "$LOGFILE" "$LOGFILE_FAILURE"
     exit 1
fi

echo -e "     nnf-ec ran successfully!"  | eval "$TEE_LOGFILE"

# Retrieve a list of unique firmware levels
firmware_list=$(ssh "$rabbit" "tools/nvme.sh cmd id-ctrl | grep '^fr '")
firmware_levels=$(echo "$firmware_list" | awk '{print $3}' | sort | uniq)
mapfile -t firmware_versions<<<"$firmware_levels"

# printf "Firmware versions detected: "
# for i in "${firmware_versions[@]}";
# do
#     printf "%s " "$i"
# done
# printf "\n"

# printf "Number of versions: %d\n" ${#firmware_versions[@]}

# At this point, if we have only 1 version of firmware present and it matches the
# expected version, we've done.
if [ "${#firmware_versions[@]}" == 1 ] && [ "${firmware_versions[0]}" == "$expectedFirmware" ]; then
     echo -e "     Drive FW is already up-to-date!"  | eval "$TEE_LOGFILE"
else
     echo -e "Firmware mismatch, downloading $firmwareFile"  | eval "$TEE_LOGFILE"

    for (( slot=1; slot <= 3; ++slot ));
    do
        # shellcheck disable=SC2029
        ssh "$rabbit" "tools/nvme.sh cmd fw-download --fw=$firmwareFile --xfer=256"

        # Action values
        # 1: Activate immediately, no reset.
        # 2: Activate after the next controller reset.
        # 3: Activate immediately and reset the controller.
        action=1
        # On the 3rd slot, we want to reset the drive controller to activate the firmware
        if ((slot == 3)); then
            action=3
        fi
        # shellcheck disable=SC2029
        ssh "$rabbit" "tools/nvme.sh cmd fw-activate --slot=$slot --action=$action"
    done
fi


#NumGudFWs=$(ssh "$rabbit" "tools/nvme.sh cmd id-ctrl | grep -e "$expectedFirmware")
declare -i NumGudFWs
NumGudFWs=$(ssh "$rabbit" tools/nvme.sh cmd id-ctrl | grep -c "$expectedFirmware")

#   echo -e "Number of drives found with latest FW is $NumGudFWs "  | tee -a $(pwd)/logs/$rabbit.log
if (( "$NumGudFWs" != 16 )); then                                                            # Should find 16 Drives
   DATE_TIME=$(date '+%Y-%m-%d %H:%M:%S')                                                 # Date/time stamp for the log file
   echo -e "\nOnly $NumGudFWs were successfully flashed to $expectedFirmware at $DATE_TIME!\n" | eval "$TEE_LOGFILE"
   cp "$LOGFILE" "$LOGFILE_FAILURE"
   exit 1
fi

echo -e "All 16 drives have the latest FW!\n" | eval "$TEE_LOGFILE"
exit 0
