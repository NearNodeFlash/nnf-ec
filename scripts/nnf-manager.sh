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

HOST=${1:-"localhost"}
PORT=${2:-"8080"}

../nnf_ec --mock --http --port $PORT &

SS="http://$HOST:$PORT/redfish/v1/StorageServices"
NNF="$SS/NNF"

curl -Ss $SS | jq
curl -SS $NNF | jq
curl --header "Content-Type: application/json" --request POST --data '{ "Capacity": { "Data": { "AllocatedBytes": 1048576 } } }' "$NNF/StoragePools" | jq

echo "$SS/NNF/StoragePools/0/CapacitySources/0/ProvidingVolumes"
curl -Ss "$SS/NNF/StoragePools/0/CapacitySources/0/ProvidingVolumes" | jq

pkill nnf_ec

