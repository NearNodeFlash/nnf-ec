#!/bin/bash

export GO_ENV="testing"

go test -v ./... > ./results.txt; cat results.txt
grep FAIL results.txt && echo "Unit tests failure" && rm results.txt && exit 1 
echo "Unit tests successful" && rm results.txt
