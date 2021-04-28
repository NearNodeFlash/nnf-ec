#!/bin/bash

echo "Running go fmt"
go fmt ./pkg/... ./cmd/... ./internal/...
if [ $? -ne 0 ] ; then
	echo "failed"
	exit 1
fi

echo "success"
exit 0