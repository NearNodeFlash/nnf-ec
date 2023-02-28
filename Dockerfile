# Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.18 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the Go source tree
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nnf-ec ./cmd/nnf_ec.go

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
FROM redhat/ubi8-minimal

WORKDIR /
COPY --from=builder /workspace/nnf-ec .
USER 65532:65532

ENTRYPOINT ["/nnf-ec"]
