#!/bin/bash

kubectl get nnfnodeecdata/ec-data -n "$1" -o json | jq .status.data > data.json
