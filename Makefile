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

.DEFAULT_GOAL := help

## Display this help message
help:
	@echo "NNF Element Controller (nnf-ec) Makefile"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Docker cleanup targets:"
	@echo "  clean-images         Remove all project-related Docker images"
	@echo "  clean-docker-all     Remove ALL unused Docker resources (aggressive)"
	@echo "  clean-docker-images  Remove unused Docker images only"
	@echo "  clean-docker-cache   Remove Docker build cache"
	@echo "  clean-project-docker Remove project Docker images and unused resources"
	@echo ""

all: image ## Build the default Docker image

vendor: ## Download Go module dependencies
	go mod vendor

vet: ## Run Go vet on all packages
	go vet `go list -f {{.Dir}} ./...`

fmt: ## Format Go source code
	go fmt `go list -f {{.Dir}} ./...`

generate: ## Generate code and redfish/swordfish message registries
	( cd ./pkg/manager-message-registry/generator && go build msgenerator.go )
	go generate ./...
	go fmt ./pkg/manager-message-registry/registries

test: ## Run Go unit tests locally
	go test -v ./...

linux: ## Build Linux binary
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${DEV_IMGNAME} ./cmd/nnf_ec.go

image: ## Build Docker image
	docker build --file Dockerfile --label $(DTR_IMGPATH):$(PROD_VERSION) --tag $(DTR_IMGPATH):$(PROD_VERSION) .

container-unit-test: ## Run containerized unit tests
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

# This and the corresponding clean-lint should go away and move to git pre-commit hook
lint: ## Run linting checks in Docker container
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

# This and the cooresponding clean-codestyle should go away and move to git pre-commit hook
codestyle: ## Run code style checks in Docker container
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION) -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

clean-lint: ## Clean lint Docker images and containers
	docker rmi $(DTR_IMGPATH)-lint:$(PROD_VERSION) || true
	docker container rm lint || true

clean-codestyle: ## Clean codestyle Docker images and containers
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true
	docker container rm codestyle || true

clean-container-unit-test: ## Clean container-unit-test Docker images and containers
	docker rmi $(DTR_IMGPATH)-container-unit-test:$(PROD_VERSION) || true
	docker container rm container-unit-test || true

# push:
# 	docker push $(DTR_IMGPATH):$(PROD_VERSION)

kind-push: image ## Load Docker image into kind cluster
	kind load docker-image $(DTR_IMGPATH):$(PROD_VERSION)

clean: clean-db ## Clean build artifacts and Docker resources
	docker container prune --force
	docker image prune --force
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION) || true
	go clean all

clean-images: ## Clean all Docker images related to this project
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-container-unit-test:$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-lint:$(PROD_VERSION) || true
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true

clean-docker-all: ## Remove ALL unused Docker resources (aggressive)
	docker system prune -a --volumes --force

clean-docker-images: ## Remove unused Docker images only
	docker image prune --force

clean-docker-cache: ## Remove Docker build cache
	docker builder prune --force

clean-project-docker: clean-images ## Remove project Docker images and unused resources
	docker container prune --force
	docker image prune --force

clean-db: ## Remove all database files
	find . -name "*.db" -type d -exec rm -rf {} +
