USER=$(shell id -un)

PROD_VERSION=$(shell sed 1q .version)
DEV_REPONAME=nnf-ec
DEV_IMGNAME=nnf-ec
DTR_IMGPATH=arti.dev.cray.com/$(DEV_REPONAME)/$(DEV_IMGNAME)

all: image

vendor:
	GOPRIVATE=stash.us.cray.com go mod vendor

fmt:
	go fmt `go list -f {{.Dir}} ./...`

image:
	docker build --rm --file Dockerfile --label $(DTR_IMGPATH):$(PROD_VERSION) --tag $(DTR_IMGPATH):$(PROD_VERSION) .

container-unit-test:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

lint:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION)-$@ -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

codestyle:
	docker build -f Dockerfile --label $(DTR_IMGPATH)-$@:$(PROD_VERSION) -t $(DTR_IMGPATH)-$@:$(PROD_VERSION) --target $@ .
	docker run --rm -t --name $@  -v $(PWD):$(PWD):rw,z $(DTR_IMGPATH)-$@:$(PROD_VERSION)

clean-lint:
	docker rmi $(DTR_IMGPATH)-lint:$(PROD_VERSION) || true

clean-codestyle:
	docker rmi $(DTR_IMGPATH)-codestyle:$(PROD_VERSION) || true

# push:
# 	docker push $(DTR_IMGPATH):$(PROD_VERSION)

kind-push:
	@echo TODO


clean:
	docker container prune --force
	docker image prune --force
	docker rmi $(DTR_IMGPATH):$(PROD_VERSION)
	go clean all
