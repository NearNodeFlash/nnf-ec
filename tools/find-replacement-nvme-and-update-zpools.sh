#!/bin/bash
#
# Script to find replacement NVMe devices for missing ZPool devices
# Matches based on size and manufacturer, prioritizing exact matches

set +e
shopt -s nullglob

usage() {
    cat <<EOF

Usage: $0 [--replace] [--dry-run] [--offline-all]
  no-args       Analyze all ZPools for missing/unavailable devices, prompting for replacements
  --replace     Automatically perform replacement of missing devices (no prompting)
  --dry-run     Show what would happen but don't make actual changes
  --offline-all Set all unavailable vdevs to OFFLINE in all zpools
  --help        Show this help message

EOF
}

# Convert various size units to MB
convert_to_mb() {
    local size=$1
    local unit
    unit=${size//[0-9.]/}
    local num
    num=${size//[^0-9.]/}

    case $unit in
        T|TB|TiB) echo "scale=2; $num * 1024 * 1024" | bc || echo "$num" ;;
        G|GB|GiB) echo "scale=2; $num * 1024" | bc || echo "$num" ;;
        M|MB|MiB) echo "scale=2; $num" | bc || echo "$num" ;;
        K|KB|KiB) echo "scale=2; $num / 1024" | bc || echo "$num" ;;
        *) echo "$num" ;;
    esac
}

# Check if size matches within tolerance
is_size_match() {
    local size_mb="$1"
    local expected_size_mb="$2"

    # Skip if expected size is not valid
    [[ -z "$expected_size_mb" || "$expected_size_mb" == "0" ]] && return 1

    # Sanitize inputs to ensure they're numeric
    size_mb=$(echo "$size_mb" | tr -cd '0-9.')
    expected_size_mb=$(echo "$expected_size_mb" | tr -cd '0-9.')

    # Compare rounded integers first (faster)
    local size_mb_int
    size_mb_int=$(echo "scale=0; ($size_mb+0.5)/1" | bc 2>/dev/null || echo "0")
    local expected_size_mb_int
    expected_size_mb_int=$(echo "scale=0; ($expected_size_mb+0.5)/1" | bc 2>/dev/null || echo "0")

    if [[ "$size_mb_int" == "$expected_size_mb_int" ]]; then
        return 0  # Match
    fi

    # Check percent difference if integers don't match
    local size_diff
    size_diff=$(echo "scale=4; $size_mb - $expected_size_mb" | bc 2>/dev/null || echo "0")
    local size_diff_abs
    size_diff_abs=$(echo "$size_diff" | tr -d '-')
    local percent_diff
    percent_diff=$(echo "scale=4; ($size_diff_abs * 100) / $expected_size_mb" | bc 2>/dev/null || echo "100")

    [[ $(echo "$percent_diff < 0.5" | bc -l 2>/dev/null || echo "0") -eq 1 ]] && return 0 || return 1
}

# Check if model matches
is_model_match() {
    local model="$1"
    local expected_model="$2"

    [[ -n "$expected_model" && "$model" == *"$expected_model"* ]] && return 0 || return 1
}

# Get most common value from associative array
get_most_common() {
    local -n values=$1
    local most_common=""
    local max_count=0

    declare -A counts
    for value in "${values[@]}"; do
        ((counts["$value"]++))
        if (( counts["$value"] > max_count )); then
            max_count=${counts["$value"]}
            most_common="$value"
        fi
    done

    echo "$most_common"
}

# List candidate devices for replacement
list_candidates() {
    local expected_size_mb="$1"
    local expected_model="$2"
    local missing_device="$3"  # New parameter to get the missing device name
    local found_candidates=0

    # Keep track of which devices have been displayed
    declare -A displayed_devices

    # Extract basename if missing_device is provided
    local missing_device_basename=""
    if [[ -n "$missing_device" && "$missing_device" =~ nvme[0-9]+n[0-9]+ ]]; then
        missing_device_basename=$(basename "$missing_device")
    else
        missing_device_basename="$missing_device"
    fi

    echo "   Candidates:"

    # First pass: check if there's a device with the same name that's offline but available
    # Only show PERFECT NAME MATCHES in the first pass
    if [[ -n "$missing_device_basename" ]]; then
        for device in "${!nvme_usage_by_device[@]}"; do
            # Check for available devices including those with ZFS states (OFFLINE, UNAVAIL, etc.)
            # Modified to consider devices that are Available with ZFS states
            [[ ! "${nvme_usage_by_device[$device]}" == "Available"* ]] && continue

            device_basename=$(basename "$device")

            # Check if this is a same-name match
            if [[ "$device_basename" == "$missing_device_basename" ]]; then
                local size_mb="${nvme_sizes_by_device[$device]}"
                local model="${nvme_models_by_device[$device]}"
                local serial="${nvme_serials_by_device[$device]}"
                local raw_device_size
                raw_device_size=$(lsblk -dn -o SIZE "$device" 2>/dev/null | tr -d '[:space:]')

                printf "   > %s: %sMB (raw: %s) | %s | %s | %s\n" \
                    "$device" "$size_mb" "$raw_device_size" "$model" "$serial" "PERFECT NAME MATCH"

                # Mark this device as displayed
                displayed_devices["$device"]=1

                ((found_candidates++))
                # Don't return immediately, continue showing other candidates
            fi

            # Also check original_name in pool
            if [[ "${nvme_device_original_name[$device]}" == "$missing_device_basename" ]]; then
                # Skip if we've already displayed this device
                [[ -n "${displayed_devices[$device]}" ]] && continue

                local size_mb="${nvme_sizes_by_device[$device]}"
                local model="${nvme_models_by_device[$device]}"
                local serial="${nvme_serials_by_device[$device]}"
                local raw_device_size
                raw_device_size=$(lsblk -dn -o SIZE "$device" 2>/dev/null | tr -d '[:space:]')

                printf "   > %s: %sMB (raw: %s) | %s | %s | %s\n" \
                    "$device" "$size_mb" "$raw_device_size" "$model" "$serial" "PERFECT NAME MATCH"

                # Mark this device as displayed
                displayed_devices["$device"]=1

                ((found_candidates++))
                # Don't return immediately, continue showing other candidates
            fi
        done
    fi

    # Second pass: check for size/model matches - only show PERFECT MATCHES (both size and model)
    # MODEL MATCH ONLY is no longer shown as a candidate
    for device in "${!nvme_usage_by_device[@]}"; do
        # Skip if not available
        [[ ! "${nvme_usage_by_device[$device]}" == "Available"* ]] && continue

        # Skip if we've already displayed this device
        [[ -n "${displayed_devices[$device]}" ]] && continue

        local size_mb="${nvme_sizes_by_device[$device]}"
        local model="${nvme_models_by_device[$device]}"
        local serial="${nvme_serials_by_device[$device]}"

        local size_match=false
        local model_match=false

        is_model_match "$model" "$expected_model" && model_match=true
        is_size_match "$size_mb" "$expected_size_mb" && size_match=true

        # Only show perfect matches (both size and model) or size-only matches
        # Don't show model-only matches
        local match_type=""
        if $size_match && $model_match; then
            match_type="PERFECT MATCH"
            ((found_candidates++))
        elif $size_match; then
            match_type="SIZE MATCH ONLY"
            ((found_candidates++))
        else
            # Skip devices that only match by model
            continue
        fi

        # Show device details
        local raw_device_size
        raw_device_size=$(lsblk -dn -o SIZE "$device" 2>/dev/null | tr -d '[:space:]')
        printf "   > %s: %sMB (raw: %s) | %s | %s | %s\n" \
            "$device" "$size_mb" "$raw_device_size" "$model" "$serial" "$match_type"

        # Mark this device as displayed
        displayed_devices["$device"]=1
    done

    [[ $found_candidates -eq 0 ]] && echo "   No suitable candidates found."
    return $found_candidates
}

# Find best matching namespace for a device
find_matching_namespace() {
    local expected_size_mb="$1"
    local expected_model="$2"

    local best_device=""
    local best_match_type=""
    local best_match_score=0  # 2=perfect match, 1=partial match

    for device in "${!nvme_usage_by_device[@]}"; do
        # Skip if not available
        [[ "${nvme_usage_by_device[$device]}" != "Available" ]] && continue

        local model="${nvme_models_by_device[$device]}"
        local size_mb="${nvme_sizes_by_device[$device]}"

        local size_match=false
        local model_match=false

        is_model_match "$model" "$expected_model" && model_match=true
        is_size_match "$size_mb" "$expected_size_mb" && size_match=true

        # Determine match quality
        local match_score=0
        if $size_match && $model_match; then
            match_score=2
        elif $size_match || $model_match; then
            match_score=1
        fi

        # Update best match if this match is better
        if [[ $match_score -gt $best_match_score ]]; then
            best_device="$device"
            best_match_type=$([[ $match_score -eq 2 ]] && echo "PERFECT MATCH" || echo "PARTIAL MATCH")
            best_match_score=$match_score
        fi
    done

    [[ -n "$best_device" ]] && echo "$best_device:$best_match_type:$best_match_score" || echo ""
}

# Scan all NVMe devices
scan_nvme_devices() {
    # Get all NVMe devices and their details
    declare -gA nvme_models_by_device
    declare -gA nvme_sizes_by_device
    declare -gA nvme_serials_by_device
    declare -gA nvme_usage_by_device
    declare -gA nvme_state_by_device
    declare -gA nvme_device_original_name  # New array to track original device name in zpool

    # Get ZPool status output once to avoid calling it for each device
    local zpool_status_output=""
    if command -v zpool &>/dev/null; then
        zpool_status_output=$(zpool status 2>/dev/null)
    fi

    echo "Scanning NVMe devices..."
    for device in /dev/nvme*n*; do
        [[ -b "$device" ]] || continue
        name=$(basename "$device")

        # Skip partitions
        [[ "$name" =~ p[0-9]+$ ]] && continue

        model=$(lsblk -dn -o MODEL "$device" 2>/dev/null | tr -d '[:space:]')
        serial=$(lsblk -dn -o SERIAL "$device" 2>/dev/null | tr -d '[:space:]')
        size=$(lsblk -dn -o SIZE "$device" 2>/dev/null | tr -d '[:space:]')
        size_mb=$(convert_to_mb "$size")

        # Initialize state and usage
        local state="ONLINE"
        usage="Available"
        local original_name=""

        # Check if device is part of a ZPool by checking both full path and basename
        device_basename=$(basename "$device")
        if [[ -n "$zpool_status_output" ]]; then
            # Check if the device appears in zpool status output
            if echo "$zpool_status_output" | grep -q "$device" ||
               echo "$zpool_status_output" | grep -q "$device_basename"; then

                # Extract the device's state from zpool status output
                device_line=$(echo "$zpool_status_output" | grep -m1 "$device_basename" || echo "")
                if [[ -n "$device_line" ]]; then
                    state=$(echo "$device_line" | awk '{print $2}')

                    # Extract pool name from device line (typically the first pool mentioned above this line)
                    original_name="$device_basename"

                    # Only mark as used if it's ONLINE, otherwise keep it available for replacement
                    if [[ "$state" == "ONLINE" ]]; then
                        usage="Used by ZFS"
                    else
                        # Device is in ZFS but not ONLINE - mark as available with state note
                        usage="Available (ZFS $state)"
                    fi
                else
                    # If we can't determine state but device is in zpool, assume it's used
                    usage="Used by ZFS"
                fi
            fi
        fi

        # If not used by ZFS, check if used by LVM or mounted
        if [[ "$usage" == "Available" || "$usage" == "Available (ZFS "* ]]; then
            if pvs "$device" &>/dev/null; then
                usage="Used by LVM"
            elif grep -q "$device" /proc/mounts; then
                usage="Mounted"
            fi
        fi

        # Also check if any partitions of this device are used by ZFS
        if [[ "$usage" == "Available" || "$usage" == "Available (ZFS "* ]] && [[ -n "$zpool_status_output" ]]; then
            for partition in "${device}p"*; do
                [[ -b "$partition" ]] || continue
                partition_basename=$(basename "$partition")

                # Check if the partition is in zpool output
                if echo "$zpool_status_output" | grep -q "$partition" ||
                   echo "$zpool_status_output" | grep -q "$partition_basename"; then

                    # Extract the partition's state from zpool status
                    partition_line=$(echo "$zpool_status_output" | grep -m1 "$partition_basename" || echo "")
                    if [[ -n "$partition_line" ]]; then
                        local partition_state
                        partition_state=$(echo "$partition_line" | awk '{print $2}')

                        # Extract pool name for partition
                        original_name="$partition_basename"

                        # Only mark as used if partition is ONLINE
                        if [[ "$partition_state" == "ONLINE" ]]; then
                            usage="Partitions used by ZFS"
                        else
                            # Partition is in ZFS but not ONLINE - keep available
                            usage="Available (ZFS partition $partition_state)"
                            state="$partition_state"  # Update device state to match partition
                        fi
                    else
                        # If we can't determine state but partition is in zpool, assume it's used
                        usage="Partitions used by ZFS"
                    fi
                    break
                fi
            done
        fi

        nvme_models_by_device["$device"]="$model"
        nvme_sizes_by_device["$device"]="$size_mb"
        nvme_serials_by_device["$device"]="$serial"
        nvme_usage_by_device["$device"]="$usage"
        # Track device state for potential future use (e.g., reporting, debugging)
        nvme_state_by_device["$device"]="$state"
        nvme_device_original_name["$device"]="$original_name"

        echo "  - Found: $device ($model, $size, $usage)"
    done
}

# Get information about ZPool devices
get_zpool_devices() {
    local pool="$1"
    local present_devices=()
    local missing_devices=()
    local numeric_devices_map=() # Store mapping of numeric vdevs to their original device paths

    # Get pool status
    local pool_status
    pool_status=$(zpool status "$pool" 2>/dev/null)
    if [[ -z "$pool_status" ]]; then
        echo "Could not get status for pool: $pool"
        return 1
    fi

    # Extract device section
    local device_section
    device_section=$(echo "$pool_status" | awk '/NAME/{flag=1; next} /errors:/{flag=0} flag')

    # First pass: Look for "was /dev/..." lines in pool status to map numeric vdevs to device paths
    while read -r line; do
        if [[ "$line" =~ was\ (/dev/[^ ]+) ]]; then
            local was_device="${BASH_REMATCH[1]}"
            # Find the vdev ID for this line by looking at the first field
            local vdev_id
            vdev_id=$(echo "$line" | awk '{print $1}')

            # Store the mapping
            numeric_devices_map+=("$vdev_id:$was_device")
            echo "DEBUG: Found vdev $vdev_id was originally $was_device" >&2
        fi
    done < <(echo "$pool_status")

    # Second pass: Parse devices and use the numeric_devices_map where needed
    while read -r line; do
        # Skip empty lines and pool name
        [[ -z "$line" || "$line" =~ ^[[:space:]]*$ || "$line" =~ ^[[:space:]]*"$pool"[[:space:]] ]] && continue

        # Extract device name (first field)
        local device
        device=$(echo "$line" | awk '{print $1}')
        [[ -z "$device" ]] && continue

        # Check if device is a special type (e.g., mirror, raidz, etc.)
        if [[ "$device" =~ ^mirror|^raidz|^spare|^log|^cache ]]; then
            continue
        fi

        # Check device state (second field)
        local state
        state=$(echo "$line" | awk '{print $2}')

        # Check if the device is a numeric ID and look up its original path
        local original_device=""
        if [[ "$device" =~ ^[0-9]+$ ]]; then
            # This is a numeric vdev ID - look it up in our map
            for mapping in "${numeric_devices_map[@]}"; do
                IFS=':' read -r id path <<< "$mapping"
                if [[ "$id" == "$device" ]]; then
                    original_device="$path"
                    break
                fi
            done

            # If we found an original device name, use it
            if [[ -n "$original_device" ]]; then
                echo "Detected numeric vdev $device was originally $original_device" >&2
                # We use the original device path but preserve the original numeric ID for replacement
                device="$device|$original_device"
            fi
        fi

        # Process device path
        local full_path_device
        if [[ "$device" == /* || "$device" == *"|"* ]]; then
            # Device already has a full path or is a composite with numeric ID and path
            full_path_device="$device"
        else
            # Device might be a short name, try to resolve with /dev/ prefix
            full_path_device="/dev/$device"
        fi

        # Check if device is physically missing or just unavailable
        local device_to_check
        if [[ "$full_path_device" == *"|"* ]]; then
            # Extract the actual device path from the composite format
            device_to_check=$(echo "$full_path_device" | cut -d'|' -f2)
        else
            device_to_check="$full_path_device"
        fi

        if [[ -b "$device_to_check" ]]; then
            # Device exists but check its state
            if [[ "$state" == "UNAVAIL" || "$state" == "OFFLINE" || "$state" == "REMOVED" ||
                  "$state" == "FAULTED" || "$state" == "DEGRADED" ]]; then
                # Device exists but is not healthy - add to missing (needs replacement)
                missing_devices+=("$full_path_device")
            else
                # Device exists and appears to be healthy
                present_devices+=("$full_path_device")
            fi
        else
            # Device physically doesn't exist or path cannot be resolved
            missing_devices+=("$full_path_device")
        fi
    done < <(echo "$device_section")

    # Sort arrays to ensure consistent device ordering
    if [[ ${#present_devices[@]} -gt 0 ]]; then
        readarray -t present_devices < <(printf '%s\n' "${present_devices[@]}" | sort -V)
    fi
    
    if [[ ${#missing_devices[@]} -gt 0 ]]; then
        readarray -t missing_devices < <(printf '%s\n' "${missing_devices[@]}" | sort -V)
    fi
    
    # Return results with clear separation between present and missing devices
    echo "${present_devices[*]}|${missing_devices[*]}"
}

# Get device size and model for ZPool devices
get_zpool_device_info() {
    local pool="$1"
    local device_info=""

    # Get pool status
    local pool_status
    pool_status=$(zpool status "$pool" 2>/dev/null)
    if [[ -z "$pool_status" ]]; then
        return 1
    fi

    # Extract device section
    local device_section
    device_section=$(echo "$pool_status" | awk '/NAME/{flag=1; next} /errors:/{flag=0} flag')

    # Get size for each device
    while read -r line; do
        # Skip empty lines and pool name
        [[ -z "$line" || "$line" =~ ^[[:space:]]*$ || "$line" =~ ^[[:space:]]*"$pool"[[:space:]] ]] && continue

        # Extract device name (first field)
        local device
        device=$(echo "$line" | awk '{print $1}')
        [[ -z "$device" ]] && continue

        # Check if device is a special type (e.g., mirror, raidz, etc.)
        if [[ "$device" =~ ^mirror|^raidz|^spare|^log|^cache ]]; then
            continue
        fi

        # Find the actual device path
        local device_path
        if [[ -b "/dev/$device" ]]; then
            device_path="/dev/$device"
        elif [[ -b "$device" ]]; then
            device_path="$device"
        else
            continue  # Skip if device doesn't exist
        fi

        # Get size and model
        local size
        size=$(lsblk -dn -o SIZE "$device_path" 2>/dev/null | tr -d '[:space:]')
        local model
        model=$(lsblk -dn -o MODEL "$device_path" 2>/dev/null | tr -d '[:space:]')

        if [[ -n "$size" ]]; then
            local size_mb
            size_mb=$(convert_to_mb "$size")
            device_info+="$device:$size_mb:$model,"
        fi
    done < <(echo "$device_section")

    # Remove trailing comma
    echo "${device_info%,}"
}

# Replace a device in a ZPool
replace_zpool_device() {
    local pool="$1"
    local old_device="$2"
    local new_device="$3"
    local force="$4"

    local replace_cmd="zpool replace"
    [[ "$force" == "true" ]] && replace_cmd="$replace_cmd -f"

    # Skip actual replacement if in dry run mode
    if $DRY_RUN; then
        replace_cmd="echo WOULD EXECUTE: $replace_cmd"
    fi

    # Handle the case where old_device is a composite of numeric ID and original path
    if [[ "$old_device" == *"|"* ]]; then
        # Split the composite device string
        local vdev_id
        vdev_id=$(echo "$old_device" | cut -d'|' -f1)
        local original_path
        original_path=$(echo "$old_device" | cut -d'|' -f2)

        echo "Replacing device in pool '$pool': using vdev ID '$vdev_id' (was $original_path) with '$new_device'"
        $replace_cmd "$pool" "$original_path" "$new_device"
    else
        # Regular case, use the device path directly
        echo "Replacing device in pool '$pool': '$old_device' with '$new_device'"
        $replace_cmd "$pool" "$old_device" "$new_device"
    fi

    return $?
}

# Analyze a ZPool for missing devices
analyze_zpool() {
    local pool="$1"

    echo -e "\nZPool: $pool"
    echo "-----------------------------------"

    # Get pool status
    local pool_status
    pool_status=$(zpool status "$pool" 2>/dev/null)
    if [[ -z "$pool_status" ]]; then
        echo "Could not get status for pool: $pool"
        return 1
    fi

    # Check if pool has errors
    if ! echo "$pool_status" | grep -q "state: ONLINE"; then
        echo "Pool is not in ONLINE state: $(echo "$pool_status" | grep "state:" | awk '{print $2}')"
    fi

    # Get devices
    local devices_info
    devices_info=$(get_zpool_devices "$pool")

    # Split into present and missing devices
    IFS='|' read -r present_devices missing_devices <<< "$devices_info"

    # Create arrays from the space-separated strings
    read -r -a present_devices_array <<< "$present_devices"
    read -r -a missing_devices_array <<< "$missing_devices"

    # Show devices
    echo "Present devices:"
    if [[ ${#present_devices_array[@]} -gt 0 ]]; then
        for device in "${present_devices_array[@]}"; do
            [[ -z "$device" ]] && continue
            echo "   - $device"
        done
    else
        echo "   None"
    fi

    echo "Missing devices:"
    if [[ ${#missing_devices_array[@]} -gt 0 ]]; then
        for device in "${missing_devices_array[@]}"; do
            [[ -z "$device" ]] && continue
            echo "   - $device"
        done
    else
        echo "   None"
        return 0
    fi

    # Get device info
    local device_info
    device_info=$(get_zpool_device_info "$pool")

    # Extract sizes and models
    declare -A present_devs_size
    declare -A present_devs_model

    IFS=',' read -r -a dev_entries <<< "$device_info"
    for entry in "${dev_entries[@]}"; do
        [[ -z "$entry" ]] && continue
        IFS=':' read -r dev size model <<< "$entry"
        present_devs_size["$dev"]="$size"
        present_devs_model["$dev"]="$model"
    done

    # Find common size and model
    local common_size
    common_size=$(get_most_common present_devs_size)
    local common_model
    common_model=$(get_most_common present_devs_model)

    echo "Using sizes collected from devices already in this ZPool:"
    for dev in "${!present_devs_size[@]}"; do
        echo "   • Device: $dev, Size: ${present_devs_size[$dev]}MB, Model: ${present_devs_model[$dev]}"
    done

    [[ -n "$common_size" ]] && echo "Common device size in ZPool '$pool': $common_size MB" ||
        echo "No common size found in existing devices in this ZPool"

    [[ -n "$common_model" ]] && echo "Common device model in ZPool '$pool': $common_model" ||
        echo "No common model found in existing devices in this ZPool"

    # Check for candidates for each missing device
    for device in "${missing_devices_array[@]}"; do
        [[ -z "$device" ]] && continue

        echo -e "\nFinding replacements for: $device"
        echo "   Expected size: ${common_size}MB, Preferred model: $common_model"

        # List candidate devices
        list_candidates "$common_size" "$common_model" "$device"
    done

    # Return status based on missing devices (for auto-replacement workflow)
    [[ ${#missing_devices_array[@]} -gt 0 ]] && return 10 || return 0
}

# Replace missing devices in a ZPool
replace_missing_devices() {
    local pool="$1"
    local replacement_count=0

    # Array to track devices already used for replacements within this pool
    declare -a used_replacement_devices

    echo -e "\nReplacing missing devices in ZPool: $pool"
    echo "-----------------------------------"

    # Get devices
    local devices_info
    devices_info=$(get_zpool_devices "$pool")

    # Split into present and missing devices
    IFS='|' read -r present_devices missing_devices <<< "$devices_info"

    # Create arrays from the space-separated strings
    read -r -a present_devices_array <<< "$present_devices"
    read -r -a missing_devices_array <<< "$missing_devices"

    if [[ ${#missing_devices_array[@]} -eq 0 ]]; then
        echo "No missing devices found in ZPool: $pool"
        return 0
    fi

    # Get device info
    local device_info
    device_info=$(get_zpool_device_info "$pool")

    # Extract sizes and models from existing devices
    declare -A present_devs_size
    declare -A present_devs_model

    IFS=',' read -r -a dev_entries <<< "$device_info"
    for entry in "${dev_entries[@]}"; do
        [[ -z "$entry" ]] && continue
        IFS=':' read -r dev size model <<< "$entry"
        present_devs_size["$dev"]="$size"
        present_devs_model["$dev"]="$model"
    done

    # Find common size and model
    local common_size=""
    local common_model=""

    # Only get common values if we have entries
    if [[ ${#present_devs_size[@]} -gt 0 ]]; then
        common_size=$(get_most_common present_devs_size)
        common_model=$(get_most_common present_devs_model)
    fi

    if [[ -n "$common_size" ]]; then
        echo "Common device size in ZPool: ${common_size}MB"
    else
        echo "No common size found"
    fi

    if [[ -n "$common_model" ]]; then
        echo "Looking for devices with model: $common_model"
    else
        echo "No common model found"
    fi

    # Process each missing device one at a time
    for missing_device in "${missing_devices_array[@]}"; do
        [[ -z "$missing_device" ]] && continue

        echo -e "\nFinding replacement for: $missing_device"

        # Debug output of already used replacement devices
        if [[ ${#used_replacement_devices[@]} -gt 0 ]]; then
            echo "Devices already used as replacements in this pool:"
            for used_dev in "${used_replacement_devices[@]}"; do
                echo "   - $used_dev"
            done
        fi

        # Filter available devices to get only offline ones
        declare -a offline_candidates=()

        # First, collect all available but offline devices
        for device in "${!nvme_usage_by_device[@]}"; do
            # Skip devices already used as replacements in this pool operation
            if [[ " ${used_replacement_devices[*]} " == *" $device "* ]]; then
                continue
            fi

            if [[ "${nvme_usage_by_device[$device]}" == "Available (ZFS "* ]]; then
                offline_candidates+=("$device")
            fi
        done

        local best_device=""
        local best_match_type=""

        # If we have offline candidates, try to find a match among them first
        if [[ ${#offline_candidates[@]} -gt 0 ]]; then
            echo "Looking at offline devices first:"
            for device in "${offline_candidates[@]}"; do
                # Skip devices already used as replacements in this pool operation
                if [[ " ${used_replacement_devices[*]} " == *" $device "* ]]; then
                    continue
                fi

                local model="${nvme_models_by_device[$device]}"
                local size_mb="${nvme_sizes_by_device[$device]}"
                local state="${nvme_state_by_device[$device]}"

                echo "   • Device: $device, Size: ${size_mb}MB, Model: $model, State: $state"

                # Try to match this offline device with the missing device
                local size_match=false
                local model_match=false

                is_model_match "$model" "$common_model" && model_match=true
                is_size_match "$size_mb" "$common_size" && size_match=true

                if $size_match && $model_match; then
                    best_device="$device"
                    best_match_type="PERFECT MATCH (OFFLINE)"
                    break
                elif $size_match || $model_match; then
                    best_device="$device"
                    best_match_type="PARTIAL MATCH (OFFLINE)"
                    # Continue in case we find a better match
                fi
            done
        fi

        # If no suitable offline device found, fall back to the regular matching
        if [[ -z "$best_device" ]]; then
            echo "No suitable offline devices found, checking other available devices..."

            # Find matching device without listing all candidates
            local best_match_score=0

            for device in "${!nvme_usage_by_device[@]}"; do
                # Skip if not available
                if [[ "${nvme_usage_by_device[$device]}" != "Available" ]]; then
                    # Debug output for skip reason
                    echo "   Skipping $device: status is ${nvme_usage_by_device[$device]}" >&2
                    continue
                fi

                # Skip devices already used as replacements in this pool operation
                if [[ " ${used_replacement_devices[*]} " == *" $device "* ]]; then
                    echo "   Skipping $device: already used in this pool" >&2
                    continue
                fi

                local model="${nvme_models_by_device[$device]}"
                local size_mb="${nvme_sizes_by_device[$device]}"
                local match_score=0

                # Check for size and model match
                is_size_match "$size_mb" "$common_size" && ((match_score++))
                is_model_match "$model" "$common_model" && ((match_score++))

                # Update best match if better
                if [[ $match_score -gt $best_match_score ]]; then
                    best_device="$device"
                    best_match_type=$([[ $match_score -eq 2 ]] && echo "PERFECT MATCH" || echo "PARTIAL MATCH")
                    best_match_score=$match_score
                fi
            done
        fi

        if [[ -n "$best_device" ]]; then
            echo "Found matching device to add: $best_device ($best_match_type)"

            # Double-check that the device isn't already used
            if [[ " ${used_replacement_devices[*]} " == *" $best_device "* ]]; then
                echo "ERROR: Device $best_device was already used in this pool operation!"
                echo "Skipping replacement for $missing_device"
                continue
            fi

            # Replace the device
            if replace_zpool_device "$pool" "$missing_device" "$best_device"; then
                ((replacement_count++))

                # Add this device to the used replacements array to prevent reuse
                used_replacement_devices+=("$best_device")
                echo "Added $best_device to list of used replacement devices for this pool"

                # Update device usage status to mark it as no longer available
                if ! $DRY_RUN; then
                    nvme_usage_by_device["$best_device"]="Used by ZFS (replacement)"
                fi
            fi
        else
            echo "   No suitable candidates found for missing device: $missing_device"
            # List available candidates for diagnostic purposes, excluding already used devices
            echo "   Remaining available candidates:"
            for device in "${!nvme_usage_by_device[@]}"; do
                # Skip if not available or already used
                if [[ "${nvme_usage_by_device[$device]}" != "Available"* ||
                      " ${used_replacement_devices[*]} " == *" $device "* ]]; then
                    continue
                fi

                local size_mb="${nvme_sizes_by_device[$device]}"
                local model="${nvme_models_by_device[$device]}"
                local serial="${nvme_serials_by_device[$device]}"

                local size_match=false
                local model_match=false

                is_model_match "$model" "$common_model" && model_match=true
                is_size_match "$size_mb" "$common_size" && size_match=true

                local match_type=""
                if $size_match && $model_match; then
                    match_type="PERFECT MATCH"
                elif $size_match; then
                    match_type="SIZE MATCH ONLY"
                elif $model_match; then
                    match_type="MODEL MATCH ONLY"
                else
                    match_type="NO MATCH"
                fi

                echo "   • $device: ${size_mb}MB | $model | $serial | $match_type"
            done
        fi
    done

    # Final message
    if [[ $replacement_count -gt 0 ]]; then
        if $DRY_RUN; then
            echo -e "\nDRY RUN: Would have replaced $replacement_count devices in ZPool $pool"
        else
            echo -e "\nStarted replacement of $replacement_count devices in ZPool $pool"
            echo "   Use 'zpool status $pool' to monitor resilver progress"
        fi
    else
        echo -e "\nNo replacements were performed"
    fi

    return 0
}

# Offline all unavailable vdevs in a zpool
offline_unavailable_vdevs() {
    local pool="$1"
    local offlining_done=false

    echo -e "\nOfflining unavailable vdevs in ZPool: $pool"
    echo "-----------------------------------"

    # Get pool status
    local pool_status
    pool_status=$(zpool status "$pool" 2>/dev/null)
    if [[ -z "$pool_status" ]]; then
        echo "Could not get status for pool: $pool"
        return 1
    fi

    # Extract device section
    local device_section
    device_section=$(echo "$pool_status" | awk '/NAME/{flag=1; next} /errors:/{flag=0} flag')

    # Process each device in the pool
    while read -r line; do
        # Skip empty lines and pool name
        [[ -z "$line" || "$line" =~ ^[[:space:]]*$ || "$line" =~ ^[[:space:]]*"$pool"[[:space:]] ]] && continue

        # Extract device name (first field)
        local device
        device=$(echo "$line" | awk '{print $1}')
        [[ -z "$device" ]] && continue

        # Check if device is a special type (e.g., mirror, raidz, etc.)
        if [[ "$device" =~ ^mirror|^raidz|^spare|^log|^cache ]]; then
            continue
        fi

        # Check device state (second field)
        local state
        state=$(echo "$line" | awk '{print $2}')

        # Only process devices that are UNAVAIL or FAULTED but not already OFFLINE
        if [[ "$state" == "UNAVAIL" || "$state" == "FAULTED" || "$state" == "DEGRADED" ]]; then
            local full_path_device
            if [[ "$device" == /* ]]; then
                # Device already has a full path
                full_path_device="$device"
            else
                # Device might be a short name, try to resolve with /dev/ prefix
                full_path_device="/dev/$device"
            fi

            # Offline the device
            if $DRY_RUN; then
                echo "DRY RUN: Would offline device '$device' in pool '$pool'"
            else
                echo "Offlining device '$device' in pool '$pool'"
                zpool offline "$pool" "$device"
                if [ $? -eq 0 ]; then
                    offlining_done=true
                else
                    echo "Failed to offline device '$device' in pool '$pool'"
                fi
            fi
        fi
    done < <(echo "$device_section")

    if $offlining_done; then
        echo "Successfully offlined unavailable vdevs in pool '$pool'"
    else
        echo "No unavailable vdevs to offline in pool '$pool'"
    fi

    return 0
}

find_replacement() {
    local pool="$1"
    local missing_device="$2"
    local expected_size_mb="$3"
    local expected_model="$4"
    local best_device=""
    local best_match_type=""
    local best_match_score=0
    local missing_device_basename=""

    # Set this to false to prevent printing the candidate list
    local silent_mode=true

    # Extract the base device name if missing_device is in nvme format
    if [[ "$missing_device" =~ nvme[0-9]+n[0-9]+ ]]; then
        missing_device_basename=$(basename "$missing_device")
    fi

    if ! $silent_mode; then
        echo "Finding replacement for missing device: $missing_device"
        echo "   Expected size: ${expected_size_mb}MB, model: $expected_model"
    fi

    # First priority: Check for devices with the same name that are OFFLINE/UNAVAIL in the pool
    # These are devices that need to be replaced but can be used as replacements themselves
    if [[ -n "$missing_device_basename" ]]; then
        for device in "${!nvme_usage_by_device[@]}"; do
            # Skip devices that are not available or already used
            if [[ ! "${nvme_usage_by_device[$device]}" == "Available"* ]]; then
                continue
            fi

            device_basename=$(basename "$device")

            # If this device has the same name as the missing device and it's marked as available
            # with a ZFS state (meaning it was an OFFLINE/UNAVAIL device that we marked available)
            if [[ "$device_basename" == "$missing_device_basename" &&
                  "${nvme_usage_by_device[$device]}" == "Available (ZFS "* ]]; then

                if ! $silent_mode; then
                    echo "   Found perfect replacement: $device"
                    echo "   This device has the same name as the missing device"
                fi

                # This is our best possible match - return immediately
                best_device="$device"
                return 0
            fi

            # If original device name in pool matches the missing device
            if [[ "${nvme_device_original_name[$device]}" == "$missing_device_basename" &&
                  "${nvme_usage_by_device[$device]}" == "Available (ZFS "* ]]; then

                if ! $silent_mode; then
                    echo "   Found perfect replacement: $device"
                    echo "   This device has the same name as the missing device in the pool"
                fi

                # This is our best possible match - return immediately
                best_device="$device"
                return 0
            fi
        done
    fi

    # Second priority: Check for devices with the same name and correct size, but aren't OFFLINE/UNAVAIL
    # These are regular available devices that happen to have the same name as the missing device
    if [[ -n "$missing_device_basename" ]]; then
        for device in "${!nvme_usage_by_device[@]}"; do
            # # Skip if not available
            # [[ "${nvme_usage_by_device[$device]}" != "Available" ]] && continue

            local device_basename=
            device_basename=$(basename "$device")
            local size_mb="${nvme_sizes_by_device[$device]}"

            # If device has the same name as the missing device and the size matches
            if [[ "$device_basename" == "$missing_device_basename" ]] && is_size_match "$size_mb" "$expected_size_mb"; then
                if ! $silent_mode; then
                    echo "   Found perfect name match with correct size: $device"
                fi

                # This is our second best possible match - return immediately
                best_device="$device"
                best_match_type="PERFECT NAME MATCH"
                best_match_score=3 # Higher score than size+model match
                return 0
            fi
        done
    fi

    # Third priority: Find a device that matches both size and model
    for device in "${!nvme_usage_by_device[@]}"; do
        # Skip if not available
        [[ ! "${nvme_usage_by_device[$device]}" == "Available"* ]] && continue

        local size_mb="${nvme_sizes_by_device[$device]}"
        local model="${nvme_models_by_device[$device]}"

        local size_match=false
        local model_match=false

        is_model_match "$model" "$expected_model" && model_match=true
        is_size_match "$size_mb" "$expected_size_mb" && size_match=true

        # Determine match quality
        local match_score=0
        if $size_match && $model_match; then
            match_score=2
        elif $size_match || $model_match; then
            match_score=1
        fi

        # Update best match if this match is better
        if [[ $match_score -gt $best_match_score ]]; then
            best_device="$device"
            best_match_type=$([[ $match_score -eq 2 ]] && echo "PERFECT MATCH" || echo "PARTIAL MATCH")
            best_match_score=$match_score
        fi
    done

    [[ -n "$best_device" ]] && echo "$best_device:$best_match_type:$best_match_score" || echo ""
}

# Main script execution

echo "===================================================="
echo "Finding replacement NVMe devices for missing ZPool devices"
echo "===================================================="

# Initialize global variables
AUTO_REPLACE=false
DRY_RUN=false
OFFLINE_ALL=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --replace)
            AUTO_REPLACE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            echo "Running in DRY RUN mode. No changes will be made."
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        --offline-all)
            OFFLINE_ALL=true
            echo "Running in OFFLINE ALL mode. All unavailable vdevs will be set to OFFLINE."
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if zpool command is available
if ! command -v zpool &>/dev/null; then
    echo "zpool command not found. ZFS tools may not be installed."
    exit 1
fi

# Scan NVMe devices
scan_nvme_devices

# Get list of ZPools
zpools_output=$(zpool list -H -o name 2>/dev/null || echo "")
if [[ -z "$zpools_output" ]]; then
    echo "No ZPools found."
    exit 0
fi

# First pass: identify ZPools with missing devices
declare -a zpools_with_missing_devices
echo -e "\nFirst pass: Scanning all ZPools for missing devices..."

while read -r pool; do
    [[ -z "$pool" ]] && continue

    # Analyze this ZPool for missing devices
    if analyze_zpool "$pool"; then
        echo "ZPool '$pool' has no missing devices"
    elif [ $? -eq 10 ]; then
        zpools_with_missing_devices+=("$pool")
    else
        echo "Error scanning ZPool: $pool"
    fi
done < <(echo "$zpools_output")

# Second pass: perform operations based on mode
if $OFFLINE_ALL; then
    echo -e "\nOfflining all unavailable vdevs in all ZPools..."

    while read -r pool; do
        [[ -z "$pool" ]] && continue
        offline_unavailable_vdevs "$pool"
    done < <(echo "$zpools_output")

    echo -e "\nOffline operation completed for all pools"
    exit 0
fi

# If not in OFFLINE_ALL mode, continue with the normal replacement workflow
if [[ ${#zpools_with_missing_devices[@]} -gt 0 ]]; then
    echo -e "\nSecond pass: Processing ${#zpools_with_missing_devices[@]} ZPools with missing devices"

    pool_count=${#zpools_with_missing_devices[@]}
    current_pool=0

    for pool in "${zpools_with_missing_devices[@]}"; do
        ((current_pool++))
        echo -e "\nProcessing ZPool: $pool for replacement ($current_pool of $pool_count)"

        if $AUTO_REPLACE; then
            echo "Auto-replacing missing devices..."
            replace_missing_devices "$pool" || echo "Error replacing devices in ZPool: $pool"
        else
            if $DRY_RUN; then
                echo "DRY RUN: Running replacement analysis..."
                replace_missing_devices "$pool" || echo "Error analyzing devices in ZPool: $pool"
            else
                echo -n "Do you want to replace missing devices in ZPool '$pool'? (y/n): "
                read -r response
                if [[ "$response" == "y" ]]; then
                    replace_missing_devices "$pool" || echo "Error replacing devices in ZPool: $pool"
                else
                    echo "Skipping replacement for ZPool '$pool'"
                fi
            fi
        fi

        # If there are more pools to process, ask user if they want to continue
        if [[ $current_pool -lt $pool_count && $AUTO_REPLACE == false && $DRY_RUN == false ]]; then
            echo -e "\nProcessed $current_pool of $pool_count ZPools with missing devices."
            echo -n "Continue to the next ZPool? (y/n): "
            read -r continue_response
            if [[ "$continue_response" != "y" ]]; then
                echo "Exiting early as requested."
                break
            fi
        fi
    done
else
    echo -e "\nNo ZPools with missing devices found"
fi

echo -e "\nScan completed"

