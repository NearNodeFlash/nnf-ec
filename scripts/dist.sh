#!/bin/bash

SERVERS=${1:-"01 02 03 04"}

for i in $SERVERS; do
    echo "Initializing NNF Server App on cn-$i"
    ssh rabbit-dev-cn-$i "pkill nnf_server"
    scp -Bq nnf_server rabbit-dev-cn-$i:/usr/bin || echo "Failed to copy nnf_server to cn-$i"
    #ssh -f rabbit-dev-cn-$1 "bash -c 'nohup /usr/bin/nnf_server > /dev/null 2>&1 &'"
done

