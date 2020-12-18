# -----------------------------------------------------------------
# Dockerfile -
#
# Provides Docker image build instructions for nnf-ec
#
# Author: Tim Morneau
#
# © Copyright 2020 Hewlett Packard Enterprise Development LP
#
# -----------------------------------------------------------------

# Copyright 2020 HPE.  All Rights Reserved
FROM dtr.dev.cray.com/baseos/centos:centos7 AS base

WORKDIR /

# Install base dependencies
#RUN yum install -y wget tar git make rpm-build
RUN yum install -y wget \
                   subversion \
                   mod_dav_svn \
                   which \
                   make \
                   gcc \
                   tar \
                   rpm-build \
                   https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm

# Get updated version of git from new repo
RUN yum install -y git

# Download and install Golang v1.14.4
RUN wget https://golang.org/dl/go1.14.4.linux-amd64.tar.gz && tar -xzf go1.14.4.linux-amd64.tar.gz && rm -rf go*.tar.gz

# Set Go environment
ENV GOROOT="/go"
ENV PATH="${PATH}:${GOROOT}/bin"  \
    LOG_LEVEL="INFO"              \
    GOPROXY="direct"              \
    GOPRIVATE="stash.us.cray.com" \
    GOSUMDB="on"                  \
    GO111MODULE="on"

ENV PROJECT ${GOPATH}/src/stash.us.cray.com/sp/nnf-ec

# Copy everything needed to create image
RUN mkdir -p $PROJECT
WORKDIR $PROJECT
COPY cmd ./cmd
COPY ec ./ec
COPY internal ./internal
COPY pkg ./pkg
COPY go.mod go.sum ./
#RUN go mod vendor
COPY vendor ./vendor

# Build nnf-ec binary
RUN set -ex && go build -v -i -o /usr/local/bin/nnf-ec ./cmd/nnf_ec.go
ENTRYPOINT ["/bin/sh"]

FROM dtr.dev.cray.com/baseos/centos:centos7 AS application
ENV PROJECT ${GOPATH}/src/stash.us.cray.com/rabsw/nnf-ec

# Pull over nnf-ec binary from base stage, add dependencies
COPY --from=0 /usr/local/bin/nnf-ec /usr/local/bin/nnf-ec
#RUN yum install -y udev bash && yum clean all

ENTRYPOINT ["/usr/local/bin/nnf-ec"]
