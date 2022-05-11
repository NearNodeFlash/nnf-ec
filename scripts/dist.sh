#!/bin/bash

# Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
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

SERVERS=${1:-"01 02 03 04"}

for i in $SERVERS; do
    echo "Initializing NNF Server App on cn-$i"
    ssh rabbit-dev-cn-$i "pkill nnf_server"
    scp -Bq nnf_server rabbit-dev-cn-$i:/usr/bin || echo "Failed to copy nnf_server to cn-$i"
    ssh -f rabbit-dev-cn-$1 "bash -c 'nohup /usr/bin/nnf_server > /dev/null 2>&1 &'"
done

