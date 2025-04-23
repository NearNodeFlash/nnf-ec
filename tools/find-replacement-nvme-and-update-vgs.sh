#!/bin/bash
#
# Script to find replacement NVMe namespaces for missing LVM PVs
# Matches based on size and manufacturer, prioritizing exact matches

set +e
shopt -s nullglob

# Convert various size units to MB
convert_to_mb() {
    local size=$1
    local unit=$(echo "$size" | sed 's/[0-9.]//g')
    local num=$(echo "$size" | sed 's/[^0-9.]//g')

    case $unit in
        T|TB) echo "scale=2; $num * 1024 * 1024" | bc || echo "$num" ;;
        G|GB) echo "scale=2; $num * 1024" | bc || echo "$num" ;;
        M|MB) echo "scale=2; $num" | bc || echo "$num" ;;
        K|KB) echo "scale=2; $num / 1024" | bc || echo "$num" ;;
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
    local size_mb_int=$(echo "scale=0; ($size_mb+0.5)/1" | bc 2>/dev/null || echo "0")
    local expected_size_mb_int=$(echo "scale=0; ($expected_size_mb+0.5)/1" | bc 2>/dev/null || echo "0")

    if [[ "$size_mb_int" == "$expected_size_mb_int" ]]; then
        return 0  # Match
    fi

    # Check percent difference if integers don't match
    local size_diff=$(echo "scale=4; $size_mb - $expected_size_mb" | bc 2>/dev/null || echo "0")
    local size_diff_abs=$(echo "$size_diff" | tr -d '-')
    local percent_diff=$(echo "scale=4; ($size_diff_abs * 100) / $expected_size_mb" | bc 2>/dev/null || echo "100")

    [[ $(echo "$percent_diff < 0.5" | bc -l 2>/dev/null || echo "0") -eq 1 ]] && return 0 || return 1
}

# Check if model matches
is_model_match() {
    local model=$1
    local expected_model=$2

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

# Find missing PVs in a volume group
get_missing_pvs() {
    local vg=$1
    local missing_pvs=()

    vgdisplay_output=$(vgdisplay -v "$vg" 2>/dev/null)
    [[ -z "$vgdisplay_output" ]] && return 1

    while read -r line; do
        [[ "$line" =~ "PV Name" ]] || continue
        pv=$(echo "$line" | awk '{print $3}')
        [[ -z "$pv" ]] && continue

        if ! [[ -b "$pv" ]]; then
            missing_pvs+=("$pv")
        fi
    done < <(echo "$vgdisplay_output" | grep "PV Name")

    echo "${missing_pvs[@]}"
}

# Get PV size from VG metadata
get_pv_size_from_metadata() {
    local vg_output=$1
    local pv=$2

    local pv_section=$(echo "$vg_output" | grep -A 20 "$pv")
    local extracted_size=$(echo "$pv_section" | grep -E "PV Size|PV Size:" | head -1 | sed -E 's/.*PV Size[^0-9]*([0-9.]+) ?([KMGT]i?B?).*/\1\2/')

    [[ -n "$extracted_size" ]] && convert_to_mb "$extracted_size" || echo "0"
}

# Find best matching namespace for a PV
find_matching_namespace() {
    local expected_size_mb=$1
    local expected_model=$2

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

# Get device information for all present PVs in a VG
get_present_pv_info() {
    local vg=$1
    local vg_output=$2
    declare -A present_pvs_size
    declare -A present_pvs_model

    while read -r line; do
        [[ "$line" =~ "PV Name" ]] || continue
        pv=$(echo "$line" | awk '{print $3}')
        [[ -z "$pv" ]] && continue

        if [[ -b "$pv" ]]; then
            # PV exists - get the underlying nvme device
            device_base=$(readlink -f "$pv" 2>/dev/null || echo "$pv")

            # Extract the model and size from the nvme device
            model="Unknown"
            raw_size=""

            for nvme_dev in "${!nvme_models_by_device[@]}"; do
                if [[ "$device_base" == "$nvme_dev"* ]]; then
                    model="${nvme_models_by_device[$nvme_dev]}"
                    raw_size=$(lsblk -dn -o SIZE "$nvme_dev" | tr -d '[:space:]')
                    break
                fi
            done

            # Get size either from raw device or LVM
            local size=""
            if [[ -n "$raw_size" ]]; then
                size="$raw_size"
            else
                size=$(pvs --noheadings -o pv_size "$pv" 2>/dev/null | tr -d '[:space:]')
                [[ -z "$size" ]] && continue
            fi

            size_mb=$(convert_to_mb "$size")
            [[ -z "$size_mb" || "$size_mb" == "0" ]] && continue

            present_pvs_size["$pv"]="$size_mb"
            present_pvs_model["$pv"]="$model"
        fi
    done < <(echo "$vg_output" | grep "PV Name")

    # Export results as a colon-separated list
    local result=""
    for pv in "${!present_pvs_size[@]}"; do
        result+="$pv:${present_pvs_size[$pv]}:${present_pvs_model[$pv]},"
    done
    echo "${result%,}"  # Remove trailing comma
}

# List candidate devices for replacement
list_candidates() {
    local expected_size_mb=$1
    local expected_model=$2
    local found_candidates=0

    echo "   Candidates:"
    for device in "${!nvme_usage_by_device[@]}"; do
        # Skip if not available
        [[ "${nvme_usage_by_device[$device]}" != "Available" ]] && continue

        local size_mb="${nvme_sizes_by_device[$device]}"
        local model="${nvme_models_by_device[$device]}"
        local serial="${nvme_serials_by_device[$device]}"

        local size_match=false
        local model_match=false

        is_model_match "$model" "$expected_model" && model_match=true
        is_size_match "$size_mb" "$expected_size_mb" && size_match=true

        # Determine match type
        local match_type=""
        if $size_match && $model_match; then
            match_type="✅ PERFECT MATCH"
            ((found_candidates++))
        elif $size_match; then
            match_type="⚠️ SIZE MATCH ONLY"
            ((found_candidates++))
        elif $model_match; then
            match_type="⚠️ MODEL MATCH ONLY"
            ((found_candidates++))
        else
            continue
        fi

        # Show device details
        local raw_device_size=$(lsblk -dn -o SIZE "$device" | tr -d '[:space:]')
        printf "   ➤ %s: %sMB (raw: %s) | %s | %s | %s\n" \
            "$device" "$size_mb" "$raw_device_size" "$model" "$serial" "$match_type"
    done

    [[ $found_candidates -eq 0 ]] && echo "   ❗ No suitable candidates found."
    return $found_candidates
}

# Replace a missing PV with a new device
replace_pv() {
    local vg=$1
    local missing_pv=$2
    local new_device=$3

    if $DRY_RUN; then
        echo "🔍 DRY RUN: Would add $new_device to VG $vg"
    else
        echo "🔄 Adding $new_device to VG $vg"
        if ! vgextend "$vg" "$new_device"; then
            echo "⚠️ Failed to add $new_device to VG $vg"
            return 1
        fi
    fi

    # Mark the device as used
    nvme_usage_by_device["$new_device"]="Used by LVM"

    if $DRY_RUN; then
        echo "🔍 DRY RUN: Would remove $missing_pv from VG $vg"
    else
        echo "➖ Removing $missing_pv from VG $vg"
        if ! vgreduce --removemissing --force "$vg"; then
            echo "⚠️ Failed to remove $missing_pv from VG $vg"
            return 1
        fi
    fi

    return 0
}

# Repair logical volumes if needed
repair_lvs() {
    local vg=$1
    [[ $DRY_RUN == true ]] && return 0

    echo -e "\n🔍 Checking for damaged LVs..."

    # List LVs in this VG
    local lvs_to_repair=()
    while read -r lv; do
        [[ -z "$lv" ]] && continue
        lvs_to_repair+=("$lv")
    done < <(lvs --noheadings -o lv_name "$vg" 2>/dev/null | tr -d ' ')

    if [[ ${#lvs_to_repair[@]} -eq 0 ]]; then
        echo "No LVs found in VG $vg"
        return 0
    fi

    echo "Found ${#lvs_to_repair[@]} LVs to check/repair"
    for lv in "${lvs_to_repair[@]}"; do
        echo "🔧 Checking LV $vg/$lv"

        # Try to activate the LV
        if ! lvchange -ay "$vg/$lv" 2>/dev/null; then
            echo "⚠️ Could not activate LV $vg/$lv"

            # Try to repair
            echo "🔧 Attempting repair of LV $vg/$lv"
            if lvconvert --repair "$vg/$lv"; then
                echo "✅ Repaired LV $vg/$lv"
                lvchange -ay "$vg/$lv"
            else
                echo "⚠️ Failed to repair LV $vg/$lv automatically"
            fi
        else
            echo "✅ LV $vg/$lv is active"
        fi
    done

    return 0
}

# Process and replace missing PVs in a volume group
replace_missing_pvs() {
    local vg=$1

    # Get VG information
    vgdisplay_output=$(vgdisplay -v "$vg" 2>/dev/null)
    if [[ -z "$vgdisplay_output" ]]; then
        echo "⚠️ Could not get VG information for $vg"
        return 1
    fi

    # Get missing PVs
    IFS=' ' read -r -a missing_pvs <<< "$(get_missing_pvs "$vg")"
    if [[ ${#missing_pvs[@]} -eq 0 ]]; then
        echo "No missing PVs found in VG '$vg'"
        return 0
    fi

    echo -e "\n🔧 Found ${#missing_pvs[@]} missing PVs in VG '$vg'"
    local replacement_count=0

    # Get information about present PVs
    local pv_info=$(get_present_pv_info "$vg" "$vgdisplay_output")

    # Extract sizes and models
    declare -A present_pvs_size
    declare -A present_pvs_model

    # Extract pv, size and model from the pv_info string
    if [[ -n "$pv_info" ]]; then
        IFS=',' read -r -a pv_entries <<< "$pv_info"
        for entry in "${pv_entries[@]}"; do
            [[ -z "$entry" ]] && continue
            # Split each entry using IFS and read into separate variables
            IFS=':' read -r pv size model <<< "$entry"
            if [[ -n "$pv" && -n "$size" ]]; then
                present_pvs_size["$pv"]="$size"
                present_pvs_model["$pv"]="$model"
            fi
        done
    fi

    # Find common size and model
    local common_size=""
    local common_model=""

    # Only get common values if we have entries
    if [[ ${#present_pvs_size[@]} -gt 0 ]]; then
        common_size=$(get_most_common present_pvs_size)
        common_model=$(get_most_common present_pvs_model)
    fi

    if [[ -n "$common_size" ]]; then
        echo "Common PV size in VG: ${common_size}MB"
    else
        echo "⚠️ No common size found"
    fi

    if [[ -n "$common_model" ]]; then
        echo "Looking for devices with model: $common_model"
    else
        echo "⚠️ No common model found"
    fi

    # Process each missing PV
    for missing_pv in "${missing_pvs[@]}"; do
        echo -e "\n🔄 Finding new PV to add before removing: $missing_pv"

        # Determine expected size
        local expected_size_mb=$(get_pv_size_from_metadata "$vgdisplay_output" "$missing_pv")
        if [[ "$expected_size_mb" == "0" && -n "$common_size" ]]; then
            expected_size_mb="$common_size"
            echo "   📏 Using common size from other PVs: ${expected_size_mb}MB"
        elif [[ "$expected_size_mb" != "0" ]]; then
            echo "   📏 Size from metadata: ${expected_size_mb}MB"
        else
            echo "   ⚠️ Could not determine size for missing PV"
        fi

        # Find best matching namespace
        local best_match=$(find_matching_namespace "$expected_size_mb" "$common_model")
        local best_device=$(echo "$best_match" | cut -d: -f1)
        local best_match_type=$(echo "$best_match" | cut -d: -f2)

        if [[ -n "$best_device" ]]; then
            echo "Found matching device to add: $best_device ($best_match_type)"

            # Replace the PV
            if replace_pv "$vg" "$missing_pv" "$best_device"; then
                ((replacement_count++))
            fi
        else
            echo "   ❗ No suitable candidates found for missing PV: $missing_pv"
            # List available candidates for diagnostic purposes
            list_candidates "$expected_size_mb" "$common_model"
        fi
    done

    # Final cleanup and repair if replacements were made
    if [[ $replacement_count -gt 0 ]]; then
        if $DRY_RUN; then
            echo -e "\n🔍 DRY RUN: Would now check and repair LVs in $vg if needed"
        else
            # Repair logical volumes if needed
            repair_lvs "$vg"
            echo -e "\n✅ Completed replacement of $replacement_count PVs in VG $vg"
        fi
    else
        echo -e "\n⚠️ No replacements were performed"
    fi

    return 0
}

# Analyze a volume group for missing PVs
analyze_vg() {
    local vg=$1

    echo -e "\n🔸 Volume Group: $vg"
    echo "-----------------------------------"

    # Get VG information
    vgdisplay_output=$(vgdisplay -v "$vg" 2>/dev/null)
    if [[ -z "$vgdisplay_output" ]]; then
        echo "⚠️ Could not get VG information for $vg"
        return 1
    fi

    # Get missing PVs
    IFS=' ' read -r -a missing_pvs <<< "$(get_missing_pvs "$vg")"

    # Get information about present PVs
    local pv_info=$(get_present_pv_info "$vg" "$vgdisplay_output")

    # Show missing PVs
    for pv in "${missing_pvs[@]}"; do
        echo "❌ MISSING PV: $pv"
    done

    if [[ ${#missing_pvs[@]} -eq 0 ]]; then
        echo "📊 No missing PVs in this volume group."
        return 0
    fi

    # Extract sizes and models
    declare -A present_pvs_size
    declare -A present_pvs_model
    IFS=',' read -r -a pv_entries <<< "$pv_info"
    for entry in "${pv_entries[@]}"; do
        [[ -z "$entry" ]] && continue
        IFS=':' read -r pv size model <<< "$entry"
        present_pvs_size["$pv"]="$size"
        present_pvs_model["$pv"]="$model"
    done

    # Find common size and model
    local common_size=$(get_most_common present_pvs_size)
    local common_model=$(get_most_common present_pvs_model)

    echo "Using sizes collected from PVs already in this VG:"
    for pv in "${!present_pvs_size[@]}"]; do
        echo "   • PV: $pv, Size: ${present_pvs_size[$pv]}MB"
    done

    [[ -n "$common_size" ]] && echo "Common PV size in VG '$vg': $common_size MB" ||
        echo "⚠️ No common size found in existing PVs in this VG"

    # Analyze each missing PV
    for missing_pv in "${missing_pvs[@]}"]; do
        echo -e "\n🔄 Finding replacements for: $missing_pv"
        echo "   Expected size from VG scan: ${common_size}MB, Preferred model: $common_model"

        # Get size from metadata
        local extracted_size=$(get_pv_size_from_metadata "$vgdisplay_output" "$missing_pv")

        # Determine expected size
        local expected_size_mb=""
        if [[ "$extracted_size" != "0" ]]; then
            expected_size_mb="$extracted_size"
            echo "   📏 Extracted size from metadata: ${expected_size_mb}MB"
        elif [[ -n "$common_size" ]]; then
            expected_size_mb="$common_size"
            echo "   📏 Using common size from other PVs: ${expected_size_mb}MB"
        else
            echo "   ⚠️ WARNING: Could not determine size for missing PV"
            expected_size_mb="0"
        fi

        echo "   📊 Using size for comparisons: ${expected_size_mb}MB"

        # List candidate devices
        list_candidates "$expected_size_mb" "$common_model"
    done

    if [[ ${#missing_pvs[@]} -gt 0 ]]; then
        echo "📋 VG '$vg' has ${#missing_pvs[@]} missing PVs that need replacement"
        return 10  # Return code 10 indicates missing PVs
    fi

    return 0
}

# Scan all NVMe devices
scan_nvme_devices() {
    # Get all NVMe devices and their details
    declare -gA nvme_models_by_device
    declare -gA nvme_sizes_by_device
    declare -gA nvme_serials_by_device
    declare -gA nvme_usage_by_device

    echo "⏳ Scanning NVMe devices..."
    for device in /dev/nvme*n*; do
        [[ -b "$device" ]] || continue
        name=$(basename "$device")

        # Skip partitions
        [[ "$name" =~ p[0-9]+$ ]] && continue

        model=$(lsblk -dn -o MODEL "$device" | tr -d '[:space:]')
        serial=$(lsblk -dn -o SERIAL "$device" | tr -d '[:space:]')
        size=$(lsblk -dn -o SIZE "$device" | tr -d '[:space:]')
        size_mb=$(convert_to_mb "$size")

        # Check if device is used by LVM or mounted
        if pvs "$device" &>/dev/null; then
            usage="Used by LVM"
        elif grep -q "$device" /proc/mounts; then
            usage="Mounted"
        else
            usage="Available"
        fi

        nvme_models_by_device["$device"]="$model"
        nvme_sizes_by_device["$device"]="$size_mb"
        nvme_serials_by_device["$device"]="$serial"
        nvme_usage_by_device["$device"]="$usage"

        echo "  - Found: $device ($model, $size, $usage)"
    done
}

# Main script execution

echo "===================================================="
echo "🔍 Finding replacement NVMe devices for missing LVM PVs"
echo "===================================================="

# Initialize global variables
AUTO_REPLACE=false
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --replace)
            AUTO_REPLACE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            echo "🔍 Running in DRY RUN mode. No changes will be made."
            shift
            ;;
        --help)
            echo "Usage: $0 [--replace] [--dry-run]"
            echo "  --replace  Automatically perform replacement of missing PVs (no prompting)"
            echo "  --dry-run  Show what would happen but don't make actual changes"
            echo "  --help     Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Scan NVMe devices
scan_nvme_devices

# Get list of volume groups
vgs_output=$(vgs --noheadings -o vg_name 2>/dev/null || echo "")
if [[ -z "$vgs_output" ]]; then
    echo "No volume groups found."
    exit 0
fi

# First pass: identify VGs with missing PVs
declare -a vgs_with_missing_pvs
echo -e "\n🔍 First pass: Scanning all volume groups for missing PVs..."

while read -r vg; do
    vg=$(echo "$vg" | tr -d '[:space:]')
    [[ -z "$vg" ]] && continue

    # Analyze this VG for missing PVs
    if analyze_vg "$vg"; then
        echo "✅ VG '$vg' has no missing PVs"
    elif [ $? -eq 10 ]; then
        vgs_with_missing_pvs+=("$vg")
    else
        echo "⚠️ Error scanning VG: $vg"
    fi
done < <(echo "$vgs_output")

# Second pass: replace missing PVs
if [[ ${#vgs_with_missing_pvs[@]} -gt 0 ]]; then
    echo -e "\n🔧 Second pass: Processing ${#vgs_with_missing_pvs[@]} VGs with missing PVs"

    for vg in "${vgs_with_missing_pvs[@]}"; do
        echo -e "\n🔸 Processing VG: $vg for replacement"

        if $AUTO_REPLACE; then
            echo "🔄 Auto-replacing missing PVs..."
            replace_missing_pvs "$vg" || echo "⚠️ Error replacing PVs in VG: $vg"
        else
            if $DRY_RUN; then
                echo "🔍 DRY RUN: Running replacement analysis..."
                replace_missing_pvs "$vg" || echo "⚠️ Error analyzing PVs in VG: $vg"
            else
                echo -n "🔄 Do you want to replace missing PVs in VG '$vg'? (y/n): "
                read -r response
                if [[ "$response" == "y" ]]; then
                    replace_missing_pvs "$vg" || echo "⚠️ Error replacing PVs in VG: $vg"
                else
                    echo "Skipping replacement for VG '$vg'"
                fi
            fi
        fi
    done
else
    echo -e "\n✅ No volume groups with missing PVs found"
fi

echo -e "\n✅ Scan completed"
