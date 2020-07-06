BIN := kube-mgmt
PKG := github.com/softonic/node-policy-webhook
REGISTRY ?= docker
VERSION := 0.0.0-dev
ARCH := amd64

IMAGE := $(REGISTRY)/$(BIN)

BUILD_IMAGE ?= golang:1.14-buster

.PHONY: all
all: image

.PHONY: build
build:
	docker run -it \
		-v $$(pwd):/go/src/$(PKG) \
		-v $$(pwd)/bin/linux_$(ARCH):/go/bin \
		-w /go/src/$(PKG) \
		$(BUILD_IMAGE) \
		/bin/sh -c "ARCH=$(ARCH) VERSION=$(VERSION) ./build/build"

.PHONY: image
image: build
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

.PHONY: clean
clean:
	rm -fr bin .go

.PHONY: undeploy
undeploy:
	kubectl delete -f manifests/deployment-opa.yml || true

.PHONY: deploy
deploy:
	kubectl create -f manifests/deployment-opa.yml

.PHONY: up
up: image undeploy deploy

.PHONY: push
push:
	docker push $(IMAGE):$(VERSION)

.PHONY: push-latest
push-latest:
	docker push $(IMAGE):latest

.PHONY: version
version:
	@echo $(VERSION)
