#!/bin/bash

# Copyright 2022-2025 Hewlett Packard Enterprise Development LP
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

usage() {
    cat <<EOF
Run various NVMe Namespace commands over the NVMe switch fabric using
switchtec-nvme utility.

Usage: $0 [-h] [-t] COMMAND [ARGS...]

Commands:
    create [SIZE-IN-BYTES]               create an nvme namespace on each drive of the specified size. (0 implies max capacity)
    attach [NAMESPACE-ID] [CONTROLLER]   attach namespaces from each drive to a controller
    delete [NAMESPACE-ID]                delete an nvme namespace on each drive
    detach [NAMESPACE-ID] [CONTROLLER]   detach namespaces from each drive for a controller
    list                                 list the nvme namespaces on each drive
    list-pdfid                           list the physical device fabric ID (PDFID) of each drive
    update-firmware [FIRMWARE-FILE] [OPTIONS]  update KIOXIA firmware on all drives

    cmd [COMMAND] [ARG [ARG]...]         execute COMMAND on each drive in the fabric.
                                         i.e. $0 id-ctrl

Arguments:
  -h                display this help
  -t                time each command

Examples:
  ./nnf-nvme.sh -t delete 1                                                 # delete namespace 1

  ./nnf-nvme.sh cmd list-ns --all                                           # list the namespaces on each drive (formerly "nnf-nvme.sh list" command)
  ./nnf-nvme.sh cmd id-ctrl | grep -E "^fr "                                # display firmware level
  ./nnf-nvme.sh cmd id-ctrl | grep -E "^mn "                                # display model name
  ./nnf-nvme.sh cmd id-ctrl | grep -e "Execute" -e "^fr " -e "^sn "         # display the drive's PDFID, firmware version, and serial number

  ./nnf-nvme.sh cmd format --force --ses=0 --namespace-id=<namespace id>    # format specified namespace
  ./nnf-nvme.sh cmd list-ctrl --namespace-id=<ns-id>                        # list the controller attached to namespace "ns-id"

  ./nnf-nvme.sh cmd virt-mgmt --cntlid=3 --act=9                            # enable virtual functions for Rabbit

KIOXIA Firmware Update:
  ./nnf-nvme.sh update-firmware /path/to/1TCRS105.std                       # update firmware on all KIOXIA drives (auto-activates)
  ./nnf-nvme.sh update-firmware /path/to/1TCRS105.std --dry-run             # simulate firmware update (dry run)
  ./nnf-nvme.sh update-firmware /path/to/1TCRS105.std --force               # force update even if not newer (auto-activates)
  ./nnf-nvme.sh update-firmware /path/to/1TCRS105.std --slot=2              # force firmware into specific slot (no auto-activate)
  ./nnf-nvme.sh update-firmware /path/to/1TCRS103.std --slot=3 --force      # force older firmware into slot 3 (no auto-activate)
  ./nnf-nvme.sh update-firmware /path/to/1TCRS105.std --dry-run --verbose   # dry run with verbose output

Drive Firmware upgrade:
  ./nnf-nvme.sh cmd fw-download --fw=<filename>.ftd                         # download firmware
  ./nnf-nvme.sh cmd fw-activate --action=3                                  # activate latest firmware download

KIOXIA Firmware upgrade (using update-firmware command):
  ./nnf-nvme.sh update-firmware <filename>.std                              # update KIOXIA firmware (recommended method - auto-activates)
  ./nnf-nvme.sh update-firmware <filename>.std --slot=2                     # force firmware into specific slot (manual activation required)

EOF
}

# executeOnSwitch <fn<path>> <switch> <args...>
executeOnSwitch() {
    local FUNCTION=$1 SWITCH=$2 ARGS=( "${@:3}" )

    # shellcheck disable=SC2086
    if [ "$(type -t $FUNCTION)" != "function" ]; then
        echo "$1 is not a function."
        exit 1
    fi

    mapfile -t PDFIDS < <(getPDFIDs "$SWITCH")
    for INDEX in "${!PDFIDS[@]}";
    do
        "$FUNCTION" "${PDFIDS[$INDEX]}@$SWITCH" "${ARGS[@]}" || echo "Error occurred for $SWITCH, continuing"
    done
}

# execute <fn<path>> <args...>
execute() {
    local FUNCTION=$1 ARGS=( "${@:2}" )

    SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
    for SWITCH in "${SWITCHES[@]}";
    do
        executeOnSwitch "$FUNCTION" "$SWITCH" "${ARGS[@]}"
    done
}

# executeParallel <fn<path>> on each switch in parallel
executeParallel() {
    local FUNCTION=$1 ARGS=( "${@:2}" )

    SWITCHES=("/dev/switchtec0" "/dev/switchtec1")
    for SWITCH in "${SWITCHES[@]}";
    do
        # To see the output as commands run use this approach. This produces
        # a mixed output that is difficult to read, but provides feedback that something is happening
        # executeOnSwitch "$FUNCTION" "$SWITCH" "${ARGS[@]}" 2>&1 | tee _result"$(basename "$SWITCH")" &

        local functionName="$FUNCTION"
        if [ "$FUNCTION" == "cmd" ];
        then
            functionName="$2"
            printf "Executing %s for each drive on %s\n" "$functionName" "$SWITCH"
        fi
        executeOnSwitch "$FUNCTION" "$SWITCH" "${ARGS[@]}" > _result"$(basename "$SWITCH")" 2>&1 &
    done
    wait

    for SWITCH in "${SWITCHES[@]}";
    do
        cat _result"$(basename "$SWITCH")"
    done

    rm _result*
}

# Retrieve the Physical Device Fabric IDs used to iterate through a list of nvme drives
function getPDFIDs() {
    local SWITCH=$1 FUNCTION="${2:-0}"

    switchtec fabric gfms-dump "$SWITCH" | grep "Function $FUNCTION " -A2 | grep PDFID | awk '{print $2}'
}

function getDriveList() {
    # DRIVES=$1
    # for DRIVE in $(ls /dev/nvme* | grep -E "nvme[[:digit:]]+$");
    for DRIVE in /dev/nvme[0-9]*;
    do
        # shellcheck disable=SC2086
        if [ "$(nvme id-ctrl ${DRIVE} | grep -e KIOXIA -e 'SAMSUNG MZ3LO1T9HCJR')" != "" ];
        then
            # SerialNumber=$(nvme id-ctrl ${DRIVE} | grep -E "^sn " | awk '{print $3}')
            # Mfg=$(nvme id-ctrl ${DRIVE} | grep -E "^mn " | awk '{print $3}')
            # FW=$(nvme id-ctrl ${DRIVE} | grep -E "^fr " | awk '{print $3}')
            # printf "%s\t%s\t%s\t%s\n" "$DRIVE" "$Mfg" "$SerialNumber" "$FW"

            DRIVES+=("${DRIVE}")
        fi
    done

    DriveCount="${#DRIVES[@]}"
    if ((DriveCount == 0));
    then
        printf "No drives found: Did you run nnf-ec?\n"
    fi
}

# ================== FIRMWARE UPDATE FUNCTIONS ==================

# Parse firmware log output from switchtec-nvme fw-log command
# Input: device firmware_log_output
# Output: slot:version pairs (one per line)
parse_firmware_log() {
    local device="$1"
    local fw_log_output="$2"

    # Extract firmware slots with versions from switchtec-nvme fw-log output
    # Format: frs1 : 0x3430315352435431 (1TCRS104)
    echo "$fw_log_output" | grep -E "^frs[1-3]" | while read -r line; do
        local pattern='^frs([1-3])[[:space:]]*:[[:space:]]*[^[:space:]]+[[:space:]]*\(([^)]+)\)'
        if [[ "$line" =~ $pattern ]]; then
            local slot="${BASH_REMATCH[1]}"
            local version="${BASH_REMATCH[2]}"
            echo "$slot:$version"
        fi
    done
}

# Get firmware information for a device
# Input: device_path
# Output: slot:version pairs (one per line)
get_firmware_info() {
    local device="$1"

    local fw_log_output
    fw_log_output=$(switchtec-nvme fw-log "$device" 2>/dev/null) || {
        echo "Failed to query firmware log for $device" >&2
        return 1
    }

    local slots=()
    mapfile -t slots < <(parse_firmware_log "$device" "$fw_log_output")

    printf '%s\n' "${slots[@]}"
}

# Compare two version strings (KIOXIA format: 1TCRS###)
# Returns: -1 if version1 < version2, 0 if equal, 1 if version1 > version2
compare_versions() {
    local version1="$1"
    local version2="$2"

    # Extract numeric part from version strings for KIOXIA firmware format (1TCRS###)
    local num1=""
    local num2=""

    # Try to extract KIOXIA format first (1TCRS###)
    if [[ "$version1" =~ 1TCRS([0-9]+) ]]; then
        num1="${BASH_REMATCH[1]}"
    else
        # Fallback to extract any numbers
        num1=$(echo "$version1" | sed 's/[^0-9]//g')
    fi

    if [[ "$version2" =~ 1TCRS([0-9]+) ]]; then
        num2="${BASH_REMATCH[1]}"
    else
        # Fallback to extract any numbers
        num2=$(echo "$version2" | sed 's/[^0-9]//g')
    fi

    # Default to 0 if no numbers found
    num1=${num1:-0}
    num2=${num2:-0}

    # Convert to integers and compare
    if [[ $((10#$num1)) -lt $((10#$num2)) ]]; then
        echo -1
    elif [[ $((10#$num1)) -gt $((10#$num2)) ]]; then
        echo 1
    else
        echo 0
    fi
}

# Find the slot with the oldest firmware version
# Input: device slot_info_array
# Output: slot number with oldest firmware
find_oldest_slot() {
    local device="$1"
    shift
    local slots=("$@")

    local oldest_slot=""
    local oldest_version=""

    for slot_info in "${slots[@]}"; do
        local slot="${slot_info%%:*}"
        local version="${slot_info##*:}"

        if [[ -z "$oldest_slot" ]] || [[ $(compare_versions "$version" "$oldest_version") -lt 0 ]]; then
            oldest_slot="$slot"
            oldest_version="$version"
        fi
    done

    echo "$oldest_slot"
}

# Check if update is needed
# Input: current_version new_version
# Returns: 0 if update needed, 1 if not needed
should_update() {
    local current_version="$1"
    local new_version="$2"

    local comparison=$(compare_versions "$current_version" "$new_version")

    if [[ "$FORCE_UPDATE" == true ]]; then
        return 0
    fi

    # Update if current version is older than new version
    [[ $comparison -lt 0 ]]
}

# Download firmware to device
# Input: device firmware_file
download_firmware() {
    local device="$1"
    local firmware_file="$2"

    echo "Downloading firmware to $device..."

    if [[ "$DRY_RUN" == true ]]; then
        echo "[DRY RUN] Would execute: switchtec-nvme fw-download --fw=\"$firmware_file\" --xfer=256 \"$device\""
        return 0
    fi

    if ! switchtec-nvme fw-download --fw="$firmware_file" --xfer=256 "$device"; then
        echo "Failed to download firmware to $device" >&2
        return 1
    fi

    echo "Firmware downloaded successfully to $device"
}

# Commit firmware to specific slot
# Input: device slot
commit_firmware() {
    local device="$1"
    local slot="$2"

    echo "Committing firmware to slot $slot on $device..."

    if [[ "$DRY_RUN" == true ]]; then
        echo "[DRY RUN] Would execute: switchtec-nvme fw-commit --slot=\"$slot\" \"$device\""
        return 0
    fi

    if ! switchtec-nvme fw-commit --slot="$slot" "$device"; then
        echo "Failed to commit firmware to slot $slot on $device" >&2
        return 1
    fi

    echo "Firmware committed successfully to slot $slot on $device"
}

# Activate firmware with action=3 (activate latest downloaded firmware)
# Input: device
activate_firmware() {
    local device="$1"

    echo "Activating firmware on $device..."

    if [[ "$DRY_RUN" == true ]]; then
        echo "[DRY RUN] Would execute: switchtec-nvme fw-activate --action=3 \"$device\""
        return 0
    fi

    if ! switchtec-nvme fw-activate --action=3 "$device"; then
        echo "Failed to activate firmware on $device" >&2
        return 1
    fi

    echo "Firmware activated successfully on $device"
}

# Update firmware for a single device
# Input: device firmware_file new_version
update_device_firmware() {
    local device="$1"
    local firmware_file="$2"
    local new_version="$3"

    echo "Processing device: $device"

    # Get firmware information
    local slots_info=()
    if ! mapfile -t slots_info < <(get_firmware_info "$device"); then
        echo "Failed to get firmware information for $device" >&2
        return 1
    fi

    if [[ ${#slots_info[@]} -eq 0 ]]; then
        echo "No firmware slots found for $device" >&2
        return 1
    fi

    echo "Found ${#slots_info[@]} firmware slots for $device:"
    for slot_info in "${slots_info[@]}"; do
        local slot="${slot_info%%:*}"
        local version="${slot_info##*:}"
        echo "  Slot $slot: $version"
    done

    local target_slot=""
    local current_version=""

    # Check if user specified a specific slot
    if [[ -n "$FORCE_SLOT" ]]; then
        # Validate the specified slot exists
        local slot_found=false
        for slot_info in "${slots_info[@]}"; do
            local slot="${slot_info%%:*}"
            local version="${slot_info##*:}"
            if [[ "$slot" == "$FORCE_SLOT" ]]; then
                target_slot="$FORCE_SLOT"
                current_version="$version"
                slot_found=true
                break
            fi
        done

        if [[ "$slot_found" == false ]]; then
            echo "Error: Specified slot $FORCE_SLOT not found on $device" >&2
            return 1
        fi

        echo "Forcing firmware into slot $target_slot (current: $current_version)"
    else
        # Find the slot with the oldest firmware (original behavior)
        local oldest_slot=$(find_oldest_slot "$device" "${slots_info[@]}")

        for slot_info in "${slots_info[@]}"; do
            local slot="${slot_info%%:*}"
            local version="${slot_info##*:}"
            if [[ "$slot" == "$oldest_slot" ]]; then
                target_slot="$oldest_slot"
                current_version="$version"
                break
            fi
        done

        echo "Oldest firmware found in slot $target_slot: $current_version"
    fi

    echo "New firmware version: $new_version"

    # Check if update is needed (unless forcing into specific slot)
    if [[ -z "$FORCE_SLOT" ]] && ! should_update "$current_version" "$new_version"; then
        if [[ "$FORCE_UPDATE" == true ]]; then
            echo "Forcing update even though current version ($current_version) is not older than new version ($new_version)"
        else
            echo "No update needed for $device. Current version ($current_version) is not older than new version ($new_version)"
            return 0
        fi
    elif [[ -n "$FORCE_SLOT" ]]; then
        echo "Forcing firmware into slot $target_slot regardless of version comparison"
    fi

    echo "Updating firmware on $device (slot $target_slot) from $current_version to $new_version"

    # Download firmware
    if ! download_firmware "$device" "$firmware_file"; then
        return 1
    fi

    # Commit firmware to the selected slot
    if ! commit_firmware "$device" "$target_slot"; then
        return 1
    fi

    # Activate firmware only if no specific slot was forced
    # When user specifies a slot, they have more control and may not want immediate activation
    if [[ -z "$FORCE_SLOT" ]]; then
        if ! activate_firmware "$device"; then
            return 1
        fi
    else
        echo "Skipping activation since specific slot was targeted (use fw-activate manually if needed)"
    fi

    echo "Successfully updated firmware on $device"
}

# Extract firmware version from filename
# Input: firmware_filename
# Output: version number
extract_firmware_version() {
    local filename="$1"
    # Extract version number from filename like "1TCRS104.std" -> "104"
    if [[ "$filename" =~ 1TCRS([0-9]+)\.std ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo "Cannot extract firmware version from filename: $filename" >&2
        return 1
    fi
}

# ============== END FIRMWARE UPDATE FUNCTIONS ==================


alias TIME=""
while getopts "th:" OPTION
do
    case "${OPTION}" in
        't')
            alias TIME=time
            export TIMEFORMAT='%3lR'
            ;;
        'h',*)
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND - 1))

# Firmware update variables (used by update-firmware command)
DRY_RUN=false
VERBOSE=false
FORCE_UPDATE=false
FORCE_SLOT=""

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

case $1 in
    create)
        function create_ns() {
            local DRIVE=$1 SIZE=$2

            if [ "$SIZE" == "0" ]; then
                echo "Reading Capacity on $DRIVE"
                SIZE=$(switchtec-nvme id-ctrl "$DRIVE" | grep tnvmcap | awk '{print $3}')
            fi

            declare -i SECTORS=$SIZE/4096
            echo "Creating Namespaces on $DRIVE with size ${SIZE}"
            TIME switchtec-nvme create-ns "$DRIVE" --nsze="$SECTORS" --ncap="$SECTORS" --block-size=4096 --nmic=1
        }
        executeParallel create_ns "${2:-0}"
        ;;
    attach)
        function attach_ns() {
            local DRIVE=$1 NAMESPACE=$2 CONTROLLER=$3
            echo "Attaching Namespace $NAMESPACE on $DRIVE to Controller $CONTROLLER"
            TIME switchtec-nvme attach-ns "$DRIVE" --namespace-id="$NAMESPACE" --controllers="$CONTROLLER"
        }
        executeParallel attach_ns "${2:-1}" "${3:-3}"
        ;;
    delete)
        function delete_ns() {
            local DRIVE=$1 NAMESPACE=$2
            echo "Deleting Namespaces $NAMESPACE on $DRIVE"
            TIME switchtec-nvme delete-ns "$DRIVE" --namespace-id="$NAMESPACE"
        }
        executeParallel delete_ns "${2:-1}"
        ;;
    detach)
        function detach_ns() {
            local DRIVE=$1 NAMESPACE=$2 CONTROLLER=$3
            echo "Detaching Namespace $NAMESPACE on $DRIVE from Controller $CONTROLLER"
            TIME switchtec-nvme detach-ns "$DRIVE" --namespace-id="$NAMESPACE" --controllers="$CONTROLLER"
        }
        executeParallel detach_ns "${2:-1}" "${3:-3}"
        ;;
    list)
        function list_ns() {
            TIME switchtec-nvme switchtec list
        }
        list_ns
        ;;
    list-pdfid)
        function list_pfid() {
            local DRIVE=$1
            echo "$DRIVE"
        }
        executeParallel list_pfid
        ;;

    cmd)
        function cmd() {
            local DRIVE=$1 CMD=$2 ARGS=( "${@:3}" )
            echo "Execute on $DRIVE $CMD" "${ARGS[@]}"
            TIME switchtec-nvme "$CMD" "$DRIVE" "${ARGS[@]}"
        }
        # execute cmd "${@:2}"
        executeParallel cmd "${@:2}"
        ;;

    update-firmware)
        # Parse firmware update options
        FIRMWARE_FILE=""
        shift # Remove 'update-firmware' from args

        # Parse firmware file and options
        while [[ $# -gt 0 ]]; do
            case $1 in
                --dry-run)
                    DRY_RUN=true
                    shift
                    ;;
                --verbose)
                    VERBOSE=true
                    shift
                    ;;
                --force)
                    FORCE_UPDATE=true
                    shift
                    ;;
                --slot=*)
                    FORCE_SLOT="${1#*=}"
                    # Validate slot number
                    if [[ ! "$FORCE_SLOT" =~ ^[1-3]$ ]]; then
                        echo "Error: Invalid slot number '$FORCE_SLOT'. Must be 1, 2, or 3." >&2
                        exit 1
                    fi
                    shift
                    ;;
                --slot)
                    if [[ $# -lt 2 ]]; then
                        echo "Error: --slot requires a slot number (1-3)" >&2
                        exit 1
                    fi
                    FORCE_SLOT="$2"
                    # Validate slot number
                    if [[ ! "$FORCE_SLOT" =~ ^[1-3]$ ]]; then
                        echo "Error: Invalid slot number '$FORCE_SLOT'. Must be 1, 2, or 3." >&2
                        exit 1
                    fi
                    shift 2
                    ;;
                -*)
                    echo "Unknown option: $1" >&2
                    echo "Valid options: --dry-run, --verbose, --force, --slot=<1-3>" >&2
                    exit 1
                    ;;
                *)
                    if [[ -z "$FIRMWARE_FILE" ]]; then
                        FIRMWARE_FILE="$1"
                    else
                        echo "Multiple firmware files specified" >&2
                        exit 1
                    fi
                    shift
                    ;;
            esac
        done

        # Validate firmware file
        if [[ -z "$FIRMWARE_FILE" ]]; then
            echo "Error: Firmware file required" >&2
            echo "Usage: $0 update-firmware <firmware-file> [--dry-run] [--verbose] [--force] [--slot=<1-3>]" >&2
            exit 1
        fi

        if [[ ! -f "$FIRMWARE_FILE" ]]; then
            echo "Error: Firmware file not found: $FIRMWARE_FILE" >&2
            exit 1
        fi

        # Extract firmware version from filename
        NEW_VERSION=$(extract_firmware_version "$(basename "$FIRMWARE_FILE")")
        if [[ $? -ne 0 ]]; then
            echo "Error: $NEW_VERSION" >&2
            exit 1
        fi

        echo "KIOXIA NVMe Drive Firmware Update"
        echo "================================="
        echo "Firmware file: $FIRMWARE_FILE"
        echo "New version: $NEW_VERSION"
        echo "Dry run: $DRY_RUN"
        echo "Verbose: $VERBOSE"
        echo "Force update: $FORCE_UPDATE"
        if [[ -n "$FORCE_SLOT" ]]; then
            echo "Force slot: $FORCE_SLOT"
        fi
        echo ""

        # Define firmware update function for executeParallel
        function update_firmware() {
            local DRIVE=$1
            echo ""
            echo "Starting firmware update for $DRIVE"

            if update_device_firmware "$DRIVE" "$FIRMWARE_FILE" "$NEW_VERSION"; then
                echo "SUCCESS: Firmware update completed for $DRIVE"
                return 0
            else
                echo "ERROR: Firmware update failed for $DRIVE" >&2
                return 1
            fi
        }

        executeParallel update_firmware
        ;;

    *)
        usage
        exit 1
        ;;
esac
