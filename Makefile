# Copyright 2020-2024 Hewlett Packard Enterprise Development LP
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

USER=$(shell id -un)

PROD_VERSION=$(shell sed 1q .version)
DEV_IMGNAME=nnf-ec
DTR_IMGPATH=ghcr.io/nearnodeflash/$(DEV_IMGNAME)

all: image

vendor:
	go mod vendor

vet:
	go vet `go list -f {{.Dir}} ./...`

fmt:
	go fmt `go list -f {{.Dir}} ./...`

generate:
	( cd ./pkg/manager-message-registry/generator && go build msgenerator.go )
	go generate ./...
	go fmt ./pkg/manager-message-registry/registries

test:
	go test -v ./...

linux:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${DEV_IMGNAME} ./cmd/nnf_ec.go

image:
	docker build --file Dockerfile --label $(DTR_IMGPATH):$(PROD_VERSION) --tag $(DTR_IMGPATH):$(PROD_VERSION) .

container-unit-test:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

# This and the corresponding clean-lint should go away and move to git pre-commit hook
lint:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

# This and the cooresponding clean-codestyle should go away and move to git pre-commit hook
codestyle:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION) -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

clean-lint:
	docker rmi $(DTR_IMGPATH)-lint:$(PROD_VERSION) || true
	docker container rm lint || true

clean-codestyle:
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true
	docker container rm codestyle || true

clean-container-unit-test:
	docker rmi $(DTR_IMGPATH)-container-unit-test:$(PROD_VERSION) || true
	docker container rm container-unit-test || true

# push:
# 	docker push $(DTR_IMGPATH):$(PROD_VERSION)

kind-push: image
	kind load docker-image $(DTR_IMGPATH):$(PROD_VERSION)

clean: clean-db
	docker container prune --force
	docker image prune --force
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION) || true
	go clean all

# Clean all Docker images related to this project
clean-images:
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-container-unit-test:$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-lint:$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true

# Aggressive Docker cleanup - removes all unused images, containers, networks, and build cache
clean-docker-all:
	docker system prune -a --volumes --force

# Clean only unused Docker images (keeps tagged images)
clean-docker-images:
	docker image prune --force

# Clean Docker build cache
clean-docker-cache:
	docker builder prune --force

# Clean everything Docker related for this project
clean-project-docker: clean-images
	docker container prune --force
	docker image prune --force

clean-db:
	find . -name "*.db" -type d -exec rm -rf {} +
