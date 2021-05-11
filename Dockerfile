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
FROM arti.dev.cray.com/baseos-docker-master-local/centos:latest AS builder

WORKDIR /

# Install base dependencies
RUN yum install -y wget \
                   which \
                   make \
                   gcc \
                   tar \
                   rpm-build \
                   https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm

# Get updated version of git from new repo
RUN yum install -y git

# Download and install Golang v1.16.3
RUN wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz && tar -xzf go1.16.3.linux-amd64.tar.gz && rm -rf go*.tar.gz

# Set Go environment
ENV GOROOT="/go"
ENV PATH="${PATH}:${GOROOT}/bin"  \
    LOG_LEVEL="INFO"              \
    GOPROXY="direct"              \
    GOPRIVATE="stash.us.cray.com" \
    GOSUMDB="on"                  \
    GO111MODULE="on"

ENV PROJECT $GOPATH/src/stash.us.cray.com/rabsw/nnf-ec

# Copy everything needed to create image
WORKDIR $PROJECT
COPY cmd ./cmd/
COPY pkg ./pkg/
COPY internal ./internal/
COPY static-analysis ./static-analysis/
COPY go.mod go.sum runContainerTest.sh ./

# vendor directory is not in git, need to generate it
RUN go mod vendor


# Build nnf-ec binary
RUN set -ex && go build -v -i -o /usr/local/bin/nnf-ec ./cmd/nnf_ec.go
ENTRYPOINT ["/bin/sh"]

# The base testing container
FROM builder AS base_testing

# go unit tests
FROM base_testing AS container-unit-test
WORKDIR $PROJECT
ENTRYPOINT ["sh", "runContainerTest.sh"]

# the static-analysis-codestyle-container
FROM base_testing as codestyle
WORKDIR $PROJECT
ENTRYPOINT ["sh", "static-analysis/docker_codestyle_entry.sh"]

# the static-analysis-lint-container
FROM base_testing as lint
WORKDIR $PROJECT
ENTRYPOINT ["sh", "static-analysis/docker_lint_entry.sh"]

# Image build is here
FROM arti.dev.cray.com/baseos-docker-master-local/centos:latest AS application
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Pull over nnf-ec binary from base stage, add dependencies
COPY --from=builder /usr/local/bin/nnf-ec /usr/local/bin/nnf-ec
#RUN yum install -y udev bash && yum clean all

ENTRYPOINT ["/usr/local/bin/nnf-ec"]
