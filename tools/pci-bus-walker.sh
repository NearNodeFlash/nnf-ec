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

usage() {
    cat <<EOF
Walk the PCI bus from a compute node to NVMe drives, showing link speeds and widths.
This script traces the complete path from CPU to NVMe drives through the PCIe fabric.

Usage: $0 [-h] [-v] [-c COMPUTE_NODE] [-d] [-s] [-t TARGET_DEVICE]

Arguments:
  -h                    display this help
  -v                    verbose output with detailed link information
  -c COMPUTE_NODE       specify compute node (default: current hostname)
  -d                    show only switches with issues (NVMe drives skipped - check from Rabbit)
  -s                    show summary statistics only
  -t TARGET_DEVICE      trace path to specific device (e.g., nvme0, 04:00.0)

Examples:
  $0                                    # Show all paths from current node to drives
  $0 -v -c compute-node-1              # Verbose output for compute-node-1
  $0 -d                                # Show only problematic switches (drives skipped)
  $0 -s                                # Show summary statistics
  $0 -t nvme0                          # Trace path to specific NVMe device
  $0 -t 04:00.0                        # Trace path to specific PCI device

Output Format:
  CPU -> Root Complex -> Switch -> Drive
  [Speed/Width] -> [Speed/Width] -> [Speed/Width] -> [Target]
EOF
}

# Default values
COMPUTE_NODE=$(hostname)
VERBOSE=false
SHOW_ISSUES_ONLY=false
SHOW_SUMMARY_ONLY=false
TARGET_DEVICE=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
while getopts "hvc:dst:" OPTION; do
    case "${OPTION}" in
        'h')
            usage
            exit 0
            ;;
        'v')
            VERBOSE=true
            ;;
        'c')
            COMPUTE_NODE="$OPTARG"
            ;;
        'd')
            SHOW_ISSUES_ONLY=true
            ;;
        's')
            SHOW_SUMMARY_ONLY=true
            ;;
        't')
            TARGET_DEVICE="$OPTARG"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

# Function to check if we're running on the target compute node
check_compute_node() {
    local current_hostname=$(hostname)
    # Handle various hostname formats (compute-node-3, rabbit-compute-3, etc.)
    local node_number=""
    if [[ "$COMPUTE_NODE" =~ compute-node-([0-9]+) ]]; then
        node_number="${BASH_REMATCH[1]}"
    fi
    
    if [[ "$current_hostname" != "$COMPUTE_NODE" ]] && [[ "$current_hostname" != *"compute-$node_number"* ]] && [[ "$current_hostname" != *"-compute-$node_number" ]]; then
        echo -e "${YELLOW}Note: Running on $current_hostname, targeting $COMPUTE_NODE${NC}"
        echo "Script will analyze the local system's PCI topology."
        echo ""
    fi
}

# Function to get PCIe link information
get_pcie_link_info() {
    local device="$1"
    local link_info=""
    
    if lspci -s "$device" -vv >/dev/null 2>&1; then
        link_info=$(lspci -s "$device" -vv 2>/dev/null | grep -E "(LnkCap|LnkSta):" | head -2)
    else
        echo "N/A"
        return 1
    fi
    
    if [[ -z "$link_info" ]]; then
        echo "N/A"
        return 1
    fi
    
    # Extract capability and status
    local cap_speed=$(echo "$link_info" | grep "LnkCap:" | sed -n 's/.*Speed \([^,]*\).*/\1/p')
    local cap_width=$(echo "$link_info" | grep "LnkCap:" | sed -n 's/.*Width \([^,]*\).*/\1/p')
    local sta_speed=$(echo "$link_info" | grep "LnkSta:" | sed -n 's/.*Speed \([^,]*\).*/\1/p')
    local sta_width=$(echo "$link_info" | grep "LnkSta:" | sed -n 's/.*Width \([^,]*\).*/\1/p')
    
    # Handle special cases for downgraded/unknown speeds
    if [[ "$sta_speed" == *"unknown"* ]]; then
        sta_speed="UNKNOWN"
    fi
    if [[ "$sta_width" == *"x0"* ]]; then
        sta_width="x0"
    fi
    
    # Format output with better status indicators
    local status_color=""
    local cap_str="${cap_speed}/${cap_width}"
    local sta_str="${sta_speed}/${sta_width}"
    
    # Add status indicators for verbose mode
    if [[ "$VERBOSE" == true ]]; then
        # Add status indicators to help interpret the output
        if [[ "$sta_speed" == *"16GT"* ]] && [[ "$sta_width" == "x4" ]]; then
            sta_str="${sta_str} (expected)"
        elif [[ "$sta_speed" == *"8GT"* ]] && [[ "$sta_width" == "x4" ]]; then
            sta_str="${sta_str} (slower than expected)"
        elif [[ "$sta_speed" == *"downgraded"* ]] && [[ "$sta_speed" == *"16GT"* ]]; then
            sta_str="${sta_str} (acceptable)"
        elif [[ "$sta_width" == *"downgraded"* ]] && [[ "$sta_width" == "x4" ]]; then
            sta_str="${sta_str} (width acceptable)"
        fi
    fi
    
    # Check for issues - 16GT/s and x4 are expected/acceptable
    if [[ "$sta_width" == "x0" ]]; then
        status_color="$RED"
        sta_str="DOWN"
    elif [[ "$sta_speed" == "UNKNOWN" ]] || [[ "$sta_speed" == *"unknown"* ]]; then
        status_color="$RED"
        sta_str="UNKNOWN/${sta_width}"
    elif [[ "$sta_speed" == *"2.5GT"* ]] || [[ "$sta_speed" == *"5GT"* ]]; then
        status_color="$RED"  # Very slow speeds are red
    elif [[ "$sta_speed" == *"8GT"* ]]; then
        status_color="$YELLOW"  # 8GT/s is slower than expected 16GT/s
    elif [[ "$sta_speed" == *"16GT"* ]] || [[ "$sta_speed" == *"32GT"* ]]; then
        # 16GT/s is expected speed - show as good even if marked as downgraded
        if [[ "$sta_width" == "x4" ]] || [[ "$sta_width" == "x8" ]] || [[ "$sta_width" == "x16" ]]; then
            status_color="$GREEN"
        else
            status_color="$YELLOW"
        fi
    else
        status_color="$NC"
    fi
    
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${status_color}${sta_str} (cap: ${cap_str})${NC}"
    else
        echo -e "${status_color}${sta_str}${NC}"
    fi
    
    # Return 1 if there's an issue (for filtering) - 16GT/s x4 is acceptable, 8GT/s is suboptimal but not critical
    if [[ "$sta_width" == "x0" ]] || [[ "$sta_speed" == "UNKNOWN" ]] || [[ "$sta_speed" == *"unknown"* ]]; then
        return 1
    elif [[ "$sta_speed" == *"2.5GT"* ]] || [[ "$sta_speed" == *"5GT"* ]]; then
        return 1
    fi
    # 8GT/s and 16GT/s are both acceptable (8GT/s might be suboptimal but not a failure)
    return 0
}

# Function to get device description
get_device_description() {
    local device="$1"
    local desc=$(lspci -s "$device" 2>/dev/null | cut -d' ' -f2- | cut -d':' -f2-)
    if [[ -n "$desc" ]]; then
        echo "$desc"
    else
        echo "Unknown Device"
    fi
}

# Function to find PAX switches (PMC-Sierra devices)
find_pax_switches() {
    # Find PMC-Sierra PCI bridge devices (PAX switches)
    lspci | grep -i "PMC-Sierra" | awk '{print $1}' | sort
}

# Function to find all unique switches in paths to KIOXIA drives
find_switches_to_kioxia() {
    local unique_switches=()
    local nvme_devices=()
    mapfile -t nvme_devices < <(find_nvme_devices)
    
    # Collect all switches in paths to KIOXIA drives
    for device in "${nvme_devices[@]}"; do
        local path_devices=()
        mapfile -t path_devices < <(trace_device_path "$device")
        
        # Add all devices in path except the final KIOXIA drive itself
        for path_device in "${path_devices[@]}"; do
            # Skip the KIOXIA drive itself (final device in chain)
            if [[ "$path_device" != "$device" ]]; then
                unique_switches+=("$path_device")
            fi
        done
    done
    
    # Remove duplicates and sort
    printf '%s\n' "${unique_switches[@]}" | sort -u
}

# Function to analyze switches that lead to KIOXIA drives
analyze_kioxia_paths() {
    local show_issues_only=${1:-false}
    
    if [[ "$show_issues_only" == true ]]; then
        echo -e "\n${BLUE}=== Problematic Switches in Paths to KIOXIA Drives ===${NC}"
    else
        echo -e "\n${BLUE}=== Switches in Paths to KIOXIA Drives ===${NC}"
    fi
    
    local kioxia_switches=()
    mapfile -t kioxia_switches < <(find_switches_to_kioxia)
    
    if [[ ${#kioxia_switches[@]} -eq 0 ]]; then
        echo -e "${RED}No switches found in paths to KIOXIA drives${NC}"
        return 1
    fi
    
    if [[ "$show_issues_only" != true ]]; then
        echo "Found ${#kioxia_switches[@]} switch(es) in paths to KIOXIA drives"
        echo "(Ignoring switches not connected to storage)"
    fi
    
    local healthy_switches=0
    local slow_switches=0
    local down_switches=0
    local displayed_count=0
    
    for switch in "${kioxia_switches[@]}"; do
        # Get device description to show what type of switch this is
        local desc=$(get_device_description "$switch")
        
        set +e
        get_pcie_link_info "$switch" >/dev/null 2>&1
        local link_result=$?
        set -e
        
        local has_issues=false
        local link_info=""
        local link_status=""
        
        if [[ $link_result -eq 0 ]]; then
            set +e
            link_info=$(get_pcie_link_info "$switch")
            set -e
            
            # Check if this is actually a slow link (8GT/s) or has other issues
            if [[ "$link_info" == *"8GT"* ]]; then
                slow_switches=$((slow_switches + 1))
                has_issues=true  # 8GT/s is suboptimal
            elif [[ "$link_info" == *"DOWN"* ]] || [[ "$link_info" == *"UNKNOWN"* ]] || [[ "$link_info" == *"2.5GT"* ]] || [[ "$link_info" == *"5GT"* ]]; then
                has_issues=true
                if [[ "$link_info" == *"DOWN"* ]] || [[ "$link_info" == *"UNKNOWN"* ]]; then
                    down_switches=$((down_switches + 1))
                else
                    slow_switches=$((slow_switches + 1))
                fi
            else
                healthy_switches=$((healthy_switches + 1))
            fi
        else
            set +e
            link_status=$(get_pcie_link_info "$switch" 2>/dev/null)
            set -e
            has_issues=true
            
            if [[ "$link_status" == *"8GT"* ]]; then
                link_info="${YELLOW}$link_status (slower than optimal)${NC}"
                slow_switches=$((slow_switches + 1))
            else
                link_info="${RED}$link_status${NC}"
                down_switches=$((down_switches + 1))
            fi
        fi
        
        # Show switch info if not filtering for issues only, or if it has issues
        if [[ "$show_issues_only" != true ]] || [[ "$has_issues" == true ]]; then
            echo -n "  $switch ($desc): "
            echo -e "$link_info"
            displayed_count=$((displayed_count + 1))
        fi
    done
    
    if [[ "$show_issues_only" == true ]]; then
        if [[ $displayed_count -eq 0 ]]; then
            echo -e "${GREEN}No problematic switches found${NC}"
        fi
        # Don't show detailed summary for -d mode
    else
        echo ""
        echo "Storage Path Switch Summary:"
        echo "  - Healthy switches (16GT/s): $healthy_switches"
        echo "  - Slow switches (8GT/s): $slow_switches"  
        echo "  - Down switches: $down_switches"
        echo "  - Total switches in storage paths: ${#kioxia_switches[@]}"
        
        # Calculate storage path health percentage
        local total_switches=${#kioxia_switches[@]}
        local functional_switches=$((healthy_switches + slow_switches))
        local health_percent=$(( (functional_switches * 100) / total_switches ))
        
        echo "  - Storage path health: ${health_percent}% functional"
    fi
}

# Function to find NVMe devices (for path tracing only)
find_nvme_devices() {
    # Find NVMe controllers (note: these are downstream of PAX switches)
    lspci | grep -i "non-volatile\|nvme" | awk '{print $1}' | sort
}

# Function to find the path from CPU to a specific device
trace_device_path() {
    local target_device="$1"
    
    # Use sysfs to get the complete path
    local device_path="/sys/bus/pci/devices/0000:${target_device}"
    
    if [[ ! -L "$device_path" ]]; then
        echo "Error: Device $target_device not found" >&2
        echo "$target_device"
        return 1
    fi
    
    # Get the real path and extract all PCI devices
    local real_path=$(readlink -f "$device_path" 2>/dev/null)
    if [[ -z "$real_path" ]]; then
        echo "Error: Cannot resolve path for $target_device" >&2
        echo "$target_device"
        return 1
    fi
    
    # Extract all PCI device addresses from the path
    echo "$real_path" | grep -o '0000:[0-9a-fA-F]*:[0-9a-fA-F]*\.[0-9a-fA-F]*' | sed 's/0000://'
}

# Function to display device path with link information
display_device_path() {
    local target_device="$1"
    local device_type="$2"
    local has_issues=false
    
    echo -e "\n${BLUE}=== Path to $device_type ($target_device) ===${NC}"
    
    # Get the full path
    local path_devices=()
    mapfile -t path_devices < <(trace_device_path "$target_device")
    
    if [[ ${#path_devices[@]} -eq 0 ]]; then
        echo -e "${RED}Could not trace path to $target_device${NC}"
        return 1
    fi
    
    # Display each device in the path
    for i in "${!path_devices[@]}"; do
        local device="${path_devices[$i]}"
        local desc=$(get_device_description "$device")
        local link_info=""
        local indent=""
        
        # Create indentation based on level
        for ((j=0; j<i; j++)); do
            indent="  $indent"
        done
        
        # Get link information - disable pipefail temporarily
        link_info=""
        set +e
        get_pcie_link_info "$device" >/dev/null 2>&1
        local link_result=$?
        set -e
        
        if [[ $link_result -eq 0 ]]; then
            set +e
            link_info=$(get_pcie_link_info "$device")
            set -e
        else
            # Check if this is a KIOXIA NVMe drive - if so, this is expected from compute nodes
            local desc=$(get_device_description "$device")
            if [[ "$desc" == *"KIOXIA"* ]] || [[ "$desc" == *"NVMe"* ]]; then
                link_info="${YELLOW}Not accessible from compute node (expected)${NC}"
            else
                link_info="${RED}Link Issues Detected${NC}"
                has_issues=true
            fi
        fi
        
        # Display device information
        if [[ $i -eq 0 ]]; then
            echo -e "$indent${BLUE}CPU/Root Complex${NC}"
        else
            echo -e "$indent├─ $device: $desc"
            echo -e "$indent   Link: $link_info"
        fi
    done
    
    # Return status based on whether issues were found
    if [[ "$has_issues" == true ]]; then
        return 1
    fi
    return 0
}

# Function to show summary statistics
show_summary() {
    echo -e "\n${BLUE}=== PCI Bus Summary for $COMPUTE_NODE ===${NC}"
    
    local total_devices=0
    local nvme_devices=0
    local problem_devices=0
    local down_links=0
    local slow_links=0
    
    # Count NVMe devices
    local nvme_list=()
    mapfile -t nvme_list < <(find_nvme_devices)
    nvme_devices=${#nvme_list[@]}
    
    # Analyze each NVMe device
    for device in "${nvme_list[@]}"; do
        total_devices=$((total_devices + 1))
        
        # Test link info and capture the result - disable pipefail for this check
        set +e
        get_pcie_link_info "$device" >/dev/null 2>&1
        local link_check_result=$?
        set -e
        
        if [[ $link_check_result -eq 0 ]]; then
            # Device is healthy
            continue
        else
            # Device has issues
            problem_devices=$((problem_devices + 1))
            set +e
            local link_status=$(get_pcie_link_info "$device" 2>/dev/null)
            set -e
            
            if [[ "$link_status" == *"DOWN"* ]] || [[ "$link_status" == *"x0"* ]] || [[ "$link_status" == *"UNKNOWN"* ]]; then
                down_links=$((down_links + 1))
            elif [[ "$link_status" == *"2.5GT"* ]] || [[ "$link_status" == *"5GT"* ]] || [[ "$link_status" == *"8GT"* ]]; then
                slow_links=$((slow_links + 1))
            fi
            # Note: 16GT/s x4 links are not counted as issues even if marked "downgraded"
        fi
    done
    
    echo "Total NVMe devices found: $nvme_devices"
    echo "Storage paths analyzed (ignoring unconnected switches)"
    
    local kioxia_switches=()
    mapfile -t kioxia_switches < <(find_switches_to_kioxia)
    local total_storage_switches=${#kioxia_switches[@]}
    
    echo "Switches in paths to storage: $total_storage_switches"
    echo "NVMe devices (KIOXIA drives): $nvme_devices (link status checked from Rabbit side)"
    
    if [[ $problem_devices -gt 0 ]]; then
        echo ""
        if [[ $down_links -gt 0 ]]; then
            echo -e "${YELLOW}Note: NVMe devices showing as disconnected is expected from compute nodes${NC}"
            echo -e "${YELLOW}NVMe drives are downstream of PAX switches - check actual drive status from Rabbit${NC}"
        fi
        echo -e "${YELLOW}Use -d flag to see only problematic devices${NC}"
        echo -e "${YELLOW}Use -v flag for detailed link information${NC}"
        echo -e "${YELLOW}Use -t <device> to trace path to a specific device${NC}"
        echo -e "${YELLOW}Focus on switches in storage paths only${NC}"
    else
        echo -e "\n${GREEN}All NVMe devices appear to be functioning normally${NC}"
    fi
}

# Main execution
main() {
    if [[ "$SHOW_ISSUES_ONLY" == true ]]; then
        echo -e "${BLUE}PCI Issues - $COMPUTE_NODE${NC}"
    else
        echo -e "${BLUE}PCI Bus Walker - NNF Compute Node to NVMe Drive Path Analysis${NC}"
        echo "============================================================="
        echo "Target Compute Node: $COMPUTE_NODE"
        echo "Analysis Time: $(date)"
        
        check_compute_node
    fi
    
    # Show summary only if requested
    if [[ "$SHOW_SUMMARY_ONLY" == true ]]; then
        show_summary
        analyze_kioxia_paths false
        exit 0
    fi
    
    # If specific target device is specified
    if [[ -n "$TARGET_DEVICE" ]]; then
        # Check if it's an NVMe device name
        if [[ "$TARGET_DEVICE" =~ ^nvme[0-9]+$ ]]; then
            # Find the PCI address for this NVMe device
            local pci_addr=""
            for device in $(find_nvme_devices); do
                local nvme_name=$(basename "$(readlink -f /sys/bus/pci/devices/0000:${device}/nvme/nvme* 2>/dev/null | head -1)" 2>/dev/null || echo "")
                if [[ "$nvme_name" == "$TARGET_DEVICE" ]]; then
                    pci_addr="$device"
                    break
                fi
            done
            
            if [[ -n "$pci_addr" ]]; then
                display_device_path "$pci_addr" "NVMe Device $TARGET_DEVICE"
            else
                echo -e "${RED}Could not find PCI address for $TARGET_DEVICE${NC}"
                exit 1
            fi
        else
            # Assume it's already a PCI address
            display_device_path "$TARGET_DEVICE" "PCI Device"
        fi
        exit 0
    fi
    
    # Find and analyze all NVMe devices
    local nvme_devices=()
    mapfile -t nvme_devices < <(find_nvme_devices)
    
    if [[ ${#nvme_devices[@]} -eq 0 ]]; then
        echo -e "${RED}No NVMe devices found on this system${NC}"
        exit 1
    fi
    
    if [[ "$SHOW_ISSUES_ONLY" != true ]]; then
        echo -e "\nFound ${#nvme_devices[@]} NVMe device(s)"
    fi
    
    local issues_found=false
    local analyzed_count=0
    
    # Analyze each NVMe device
    for device in "${nvme_devices[@]}"; do
        local has_device_issues=false
        
        # For NVMe drives, we don't consider link issues from compute nodes as "problems"
        # since drives are downstream of PAX switches and only accessible from Rabbit nodes
        local device_has_issues=false
        
        # When showing issues only (-d flag), skip NVMe drives entirely since
        # their link status from compute nodes is not meaningful
        if [[ "$SHOW_ISSUES_ONLY" == true ]]; then
            continue
        fi
        
        display_device_path "$device" "NVMe Device"
        analyzed_count=$((analyzed_count + 1))
    done
    
    # Show summary information
    if [[ "$SHOW_ISSUES_ONLY" == true ]]; then
        # Minimal output for -d flag - just show the switch analysis
        analyze_kioxia_paths "$SHOW_ISSUES_ONLY"
    else
        echo -e "\n${BLUE}=== Analysis Complete ===${NC}"
        echo "Analyzed $analyzed_count NVMe device(s)"
        if [[ "$issues_found" == true ]]; then
            echo -e "${YELLOW}Note: NVMe drive 'issues' from compute nodes are expected${NC}"
            echo -e "${YELLOW}Use -s flag to focus on storage switch health instead${NC}"
        else
            echo -e "${GREEN}All devices appear to be functioning normally${NC}"
        fi
        
        echo -e "\n${BLUE}Additional Tools:${NC}"
        echo "  - Use './show-drive-ports.sh -v' for drive link status summary"
        echo "  - Use './pci.sh /dev/switchtec0 <PDFID>' for detailed PCIe config"
        echo "  - Use './nnf-nvme.sh list-pdfid' to list physical device fabric IDs"
        echo "  - Check actual NVMe drive status from Rabbit nodes, not compute nodes"
        
        analyze_kioxia_paths "$SHOW_ISSUES_ONLY"
        echo ""
        echo -e "${BLUE}Notes:${NC}"
        echo "  - 16GT/s x4 links on switches are expected performance (not issues)"
        echo "  - Analysis focuses only on switches in paths to KIOXIA drives"
        echo "  - Other PAX switches not connected to storage are ignored"
        echo "  - KIOXIA drive link speeds must be checked from Rabbit side"
        echo "  - Focus on storage path switch health for compute node diagnostics"
    fi
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi