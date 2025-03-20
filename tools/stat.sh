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

shopt -s expand_aliases

usage() {
    cat <<EOF
Survey the Rabbits to determine our status.
Usage: $0 [-h] [-v verbose]

Arguments:
  -v                verbose
  -h                display this help
EOF
}

divider() {
    printf "++++++++++++++++++++++++++++++++++++++++++++\n"
}

verbose=false

while getopts ":vh" OPTION
do
    case "${OPTION}" in
        'v')
            verbose=true
            ;;
        'h')
            usage
            exit 0
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND - 1))


# SYSTEM will be "elcap" or "tuolumne" based on our node's hostname
SYSTEM=${1:-$(hostname | awk '{gsub(/[0-9]+/, "", $1); print $1}')}

if [ "$SYSTEM" == "elcap" ]
then
    RABBITID="201-896"
    export R=eelcap["$RABBITID"]
    export RnoE=elcap["$RABBITID"]
    export S=elcap-jbod["$RABBITID"]
else
    RABBITID="201-272"
    export R=etuolumne["$RABBITID"]
    export RnoE=tuolumne["$RABBITID"]
    export S=tuolumne-jbod["$RABBITID"]
fi

# export C=$(for i in $(nodeset -e $R); do xhost-query.py $(xhost-query.py $i | cut -c -7)* | awk '/nodes:/ {print $2}'; done | nodeset -f)
NODES=$(kubectl get nodes --no-headers)
RABBITS_READY_IN_K8s=$(echo "$NODES" | grep -vE "${SYSTEM}[0-9]{2}\b" | grep -v NotReady | awk '{print $1}' | nodeset -f)
if $verbose; then
    printf "Rabbits ready in K8s\n"
    printf " %s (%d) \n" "$RABBITS_READY_IN_K8s" "$(cluset -c "$RABBITS_READY_IN_K8s")"
fi

RABBITS_NOT_READY_OR_NOT_IN_K8s=$(nodeset -f "$RnoE" -x "$RABBITS_READY_IN_K8s")
if $verbose; then
    printf "Rabbits not ready or not in K8s\n"
    printf " %s (%d) \n" "$RABBITS_NOT_READY_OR_NOT_IN_K8s" "$(cluset -c "$RABBITS_NOT_READY_OR_NOT_IN_K8s")"
fi

STORAGES=$(kubectl get storages --no-headers)
RABBITS_ENABLED_IN_K8s=$(echo "$STORAGES" | grep Enabled | awk '{print $1}' | nodeset -f)
if $verbose; then
    printf "Rabbits enabled in K8s\n"
    printf " %s (%d) \n" "$RABBITS_ENABLED_IN_K8s" "$(cluset -c "$RABBITS_ENABLED_IN_K8s")"
fi

RABBITS_DISABLED_IN_K8s=$(echo "$STORAGES" | grep Disabled | awk '{print $1}' | nodeset -f)
if $verbose; then
    printf "Rabbits disabled in K8s\n"
    printf " %s (%d) \n" "$RABBITS_DISABLED_IN_K8s" "$(cluset -c "$RABBITS_DISABLED_IN_K8s")"
fi

RABBITS_ENABLED_AND_READY_K8s=$(nodeset -f "$RABBITS_READY_IN_K8s" -i "$RABBITS_ENABLED_IN_K8s")
if $verbose; then
    printf "Rabbits ready and enabled in K8s\n"
    printf " %s (%d) \n" "$RABBITS_ENABLED_AND_READY_K8s" "$(cluset -c "$RABBITS_ENABLED_AND_READY_K8s")"
fi

RABBITS_DISABLED_AND_READY_K8s=$(nodeset -f "$RABBITS_READY_IN_K8s" -i "$RABBITS_DISABLED_IN_K8s")
if $verbose; then
    printf "Rabbits ready and disabled in K8s\n"
    printf " %s (%d) \n" "$RABBITS_DISABLED_AND_READY_K8s" "$(cluset -c "$RABBITS_DISABLED_AND_READY_K8s")"
fi

RABBITS_NOT_READY_FOR_JOBS=$(nodeset -f "$RABBITS_DISABLED_IN_K8s" "$RABBITS_NOT_READY_OR_NOT_IN_K8s")
if $verbose; then
    printf "Rabbits not ready for jobs\n"
    printf " %s (%d) \n" "$RABBITS_NOT_READY_FOR_JOBS" "$(cluset -c "$RABBITS_NOT_READY_FOR_JOBS")"
fi

NODEDIAG_NDRP_FOR_RABBITS_ENABLED_AND_READY_IN_K8s=$(sudo cmd_wrapper ndrp "$RABBITS_ENABLED_AND_READY_K8s" 2>/dev/null)

NODEDIAG_NDRP_FOR_RABBITS_DISABLED_AND_READY_IN_K8s=$(sudo cmd_wrapper ndrp "$RABBITS_DISABLED_AND_READY_K8s" 2>/dev/null)

RABBITS_ENABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP=$(echo "$NODEDIAG_NDRP_FOR_RABBITS_ENABLED_AND_READY_IN_K8s" | grep OK | awk '{print $1}' | sed 's/://' | cluset -f)

RABBITS_DISABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP=$(echo "$NODEDIAG_NDRP_FOR_RABBITS_DISABLED_AND_READY_IN_K8s" | grep OK | awk '{print $1}' | sed 's/://' | cluset -f)

RABBITS_ENABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP=$(echo "$NODEDIAG_NDRP_FOR_RABBITS_ENABLED_AND_READY_IN_K8s" | grep FAIL | awk '{print $1}' | sed 's/://' | cluset -f)

RABBITS_DISABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP=$(echo "$NODEDIAG_NDRP_FOR_RABBITS_DISABLED_AND_READY_IN_K8s" | grep FAIL | awk '{print $1}' | sed 's/://' | cluset -f)

RABBITS_RUNNING_ETHERNET=$(sudo cmd_wrapper isr "$R" 2>/dev/null | grep degraded | awk '{print $1}' | sed 's/://' | cluset -f)

RABBITS_NOT_RUNNING_ETHERNET=$(cluset -f "$R" -x "$RABBITS_RUNNING_ETHERNET")
RABBITS_NOT_RUNNING_ETHERNET_NO_E=${RABBITS_NOT_RUNNING_ETHERNET/^e/}

RABBITS_NOT_READY_OR_NOT_IN_K8s_BUT_RUNNING=$(nodeset -f "$RABBITS_NOT_READY_OR_NOT_IN_K8s" -x "$RABBITS_NOT_RUNNING_ETHERNET_NO_E")


# Numbers first, always numbers
printf "SYSTEM %s\n" "$SYSTEM"
printf "Rabbits ready for jobs                                                = %d\n" "$(cluset -c "$RABBITS_ENABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP")"
printf "Rabbits not ready for jobs                                            = %d\n" "$(cluset -c "$RABBITS_NOT_READY_FOR_JOBS")"
printf "    Rabbits disabled                                                  = %d\n" "$(cluset -c "$RABBITS_DISABLED_IN_K8s")"
printf "    Rabbits not ready or not in k8s, (kubernetes issue)               = %d\n" "$(cluset -c "$RABBITS_NOT_READY_OR_NOT_IN_K8s_BUT_RUNNING")"
printf "    Rabbits disabled, passing 'nodediag_wrapper' (return to service?) = %d\n" "$(cluset -c "$RABBITS_DISABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP")"
printf "    Rabbits enabled in K8s, failing 'ndrp'                            = %d\n" "$(cluset -c "$RABBITS_ENABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP")"
printf "    Rabbits not running   (OS not reachable via management network)   = %d\n" "$(cluset -c "$RABBITS_NOT_RUNNING_ETHERNET")"

divider

printf "Rabbits ready for flux jobs (enabled AND passing nodediag)\n"
printf " %s (%d)\n" "$RABBITS_ENABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP" "$(cluset -c "$RABBITS_ENABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP")"

divider

printf "Rabbits not ready for flux jobs\n"
printf " %s (%d)\n" "$RABBITS_NOT_READY_FOR_JOBS" "$(cluset -c "$RABBITS_NOT_READY_FOR_JOBS")"

divider

printf "Rabbits disabled\n"
printf " %s (%d)\n" "$RABBITS_DISABLED_IN_K8s" "$(cluset -c "$RABBITS_DISABLED_IN_K8s")"

divider

printf "Rabbits not ready or not in K8s (kubernetes issue)\n"
printf " %s (%d)\n" "$RABBITS_NOT_READY_OR_NOT_IN_K8s_BUT_RUNNING" "$(cluset -c "$RABBITS_NOT_READY_OR_NOT_IN_K8s_BUT_RUNNING")"

divider

printf "Rabbits disabled, passing 'nodediag_wrapper' (return to service?)\n"
printf " %s (%d)\n" "$RABBITS_DISABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP" "$(cluset -c "$RABBITS_DISABLED_AND_READY_IN_K8s_PASSING_NODEDIAG_RBP")"

divider

printf "Rabbits enabled but failing 'nodediag ndrp' (Disable these)\n"
printf " %s (%d)\n" "$RABBITS_ENABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP" "$(cluset -c "$RABBITS_ENABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP")"

divider

printf "Rabbits disabled and failing 'nodediag ndrp' (Look at these)\n"
printf " %s (%d)\n" "$RABBITS_DISABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP" "$(cluset -c "$RABBITS_DISABLED_AND_READY_IN_K8s_FAILING_NODEDIAG_NDRP")"

divider

printf "Rabbits 'is-system-running=degraded' on management network\n"
printf " %s (%d)\n" "$RABBITS_RUNNING_ETHERNET" "$(cluset -c "$RABBITS_RUNNING_ETHERNET")"
printf "Rabbits not running, i.e. OS not reachable via management network\n"
printf " %s (%d)\n" "$RABBITS_NOT_RUNNING_ETHERNET" "$(cluset -c "$RABBITS_NOT_RUNNING_ETHERNET")"
