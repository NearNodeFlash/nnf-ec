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

USER=$(shell id -un)

PROD_VERSION=$(shell sed 1q .version)
DEV_REPONAME=nnf-ec
DEV_IMGNAME=nnf-ec
DTR_IMGPATH=arti.dev.cray.com/$(DEV_REPONAME)/$(DEV_IMGNAME)

all: image

vendor:
	GOPRIVATE=github.hpe.com go mod vendor

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
	env GOOS=linux GOARCH=amd64 GOPRIVATE=github.com go build -o ${DEV_IMGNAME} ./cmd/nnf_ec.go

image:
	docker build --rm --file Dockerfile --label $(DTR_IMGPATH):$(PROD_VERSION) --tag $(DTR_IMGPATH):$(PROD_VERSION) .

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

clean-codestyle:
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true

# push:
# 	docker push $(DTR_IMGPATH):$(PROD_VERSION)

kind-push: image
	kind load docker-image $(DTR_IMGPATH):$(PROD_VERSION)

clean: clean-db
	docker container prune --force
	docker image prune --force
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION)
	go clean all

clean-db:
	find . -name "*.db" -type d -exec rm -rf {} +