#!/bin/bash

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

