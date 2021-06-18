# -----------------------------------------------------------------
# Dockerfile -
#
# Provides Docker image build instructions for nnf-ec
#
# Author: Nate Roiger
#
# © Copyright 2021 Hewlett Packard Enterprise Development LP
#
# -----------------------------------------------------------------

# Copyright 2020 HPE.  All Rights Reserved
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the Go source tree
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on GOPRIVATE=stash.us.cray.com go build -a -o nnf-ec ./cmd/nnf_ec.go

# Run Go unit tests
FROM builder AS container-unit-test
COPY runContainerTest.sh runContainerTest.sh
ENTRYPOINT ["sh", "runContainerTest.sh"]

# Setup Static Analysis
# TODO: These should move to pre-commit check
FROM builder as codestyle
COPY static-analysis/docker_codestyle_entry.sh static-analysis/docker_codestyle_entry.sh
ENTRYPOINT ["sh", "static-analysis/docker_codestyle_entry.sh"]

# the static-analysis-lint-container
FROM builder as lint
COPY static-analysis/docker_lint_entry.sh static-analysis/docker_lint_entry.sh
ENTRYPOINT ["sh", "static-analysis/docker_lint_entry.sh"]

# The final application release product container
FROM arti.dev.cray.com/baseos-docker-master-local/centos:latest

WORKDIR /
COPY --from=builder /workspace/nnf-ec .
USER 65532:65532

ENTRYPOINT ["/nnf-ec"]
