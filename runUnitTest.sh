#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

set -e
trap 'catch $? $LINENO' EXIT

catch() {
    if [ "$1" != "0" ]; then
        echo "Error $1 occurred on line $2"
    fi
}

# We need to run the unit test within the docker container.
# If we make the container-unit-test, we will use `docker build` to run the unit test
# within the container.
make container-unit-test