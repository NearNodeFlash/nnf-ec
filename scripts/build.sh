#!/bin/bash

set -e
trap 'catch $? $LINENO' EXIT

catch() {
    if [ "$1" != "0" ]; then
        echo "Error $1 occurred on line $2"
    fi
}

fn=nnf-ec
docker build --rm --file Dockerfile --label dtr.dev.cray.com/$USER/$fn:latest --tag dtr.dev.cray.com/$USER/$fn:latest .
docker push dtr.dev.cray.com/$USER/${fn}:latest