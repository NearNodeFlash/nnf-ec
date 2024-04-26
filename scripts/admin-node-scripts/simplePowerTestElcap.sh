#!/bin/bash

# Copyright 2024 Hewlett Packard Enterprise Development LP
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
Powercycle the compute nodes in the specified cabinet

Usage: $0 [-h] [-t] [compute-node-node-controllers]

X-NAMES:
    # Texas TDS systems
    # Rabbit-p names
    x9000c[0-7]             Chassis 0..7, compute slots 0..7, boards 0..1, node 0

Arguments:
  -h                display this help
  -t                time each command

Examples:
  # Texas TDS
  ./simplePowerTest.sh -t x9000c1
EOF
}

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

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

cabinet="${1:-x1606}"
NODES="8681-8808"
computeNodes="elcap"["$NODES"]
rabbitS="pelcap[681-688]-s"


# clear out the results file
true > results.out
true > assertCheck

delay=60
loopCount=0
while true;
do
    # Power cycle compute nodes
    ./powercycleElcap.sh -t "$computeNodes"

    pdsh -w "$rabbitS" "grep -e WDG -e ASSERT pax[0-1].log" >> assertCheck

    assertlines=$(wc -l assertCheck | awk '{print $1}')
    if [ "$assertlines" -ne 0 ]
    then
        printf "ASSERT found\n"
        cat assertCheck
        cat results.out
        exit 1
    fi

    loopCount=$((loopCount + 1))

    echo "                                                          PowerCycle count" "$loopCount" | tee -a results.out
done
