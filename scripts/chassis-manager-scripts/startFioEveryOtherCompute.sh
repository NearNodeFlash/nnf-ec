#!/bin/bash

# Copyright 2023 Hewlett Packard Enterprise Development LP
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
# set -e
# set -o xtrace

computeNodes=(
    x9000c3s0b0n0
    # x9000c3s0b1n0
    x9000c3s1b0n0
    # x9000c3s1b1n0
    x9000c3s2b0n0
    # x9000c3s2b1n0
    x9000c3s3b0n0
    # x9000c3s3b1n0
    x9000c3s4b0n0
    # x9000c3s4b1n0
    x9000c3s5b0n0
    # x9000c3s5b1n0
    x9000c3s6b0n0
    # x9000c3s6b1n0
    x9000c3s7b0n0
    # x9000c3s7b1n0
)


printf "rsync .fio configurations\n"
for nC in "${computeNodes[@]}";
do
    rsync *.fio "$nC":
done

fioFiles=(*.fio)

for file in "${fioFiles[@]}";
do
    operation=$(cat $file | grep rw | sed 's|rw=||g')
    printf "  fio operation %s\n" "$operation"

    for nC in "${computeNodes[@]}";
    do
        printf "Compute node %s:\n" "$nC"
        ssh -f "$nC" "rm -f results.$operation.* && nohup fio $file > results.$operation.$nC &"

        # The following section of commands works only if the fio version on the client matches the fio on the server!!!
        # The following commands setup an fio server on each BP node
        # ssh -f "$nC" "fio --server"
        # This command launches fio on each BP node sending them the randomwrite.fio file
        # fio --client="$nC" randomwrite.fio

    done

    sleepTime=$(( $(cat $file | grep runtime | sed 's|runtime=||g') + 5 ))
    printf "  sleeping %d seconds for $operation to finish\n" "$sleepTime"
    sleep "$sleepTime"
done

printf "\n"

./collectFioResultsEveryOther.sh
printf "\n"
