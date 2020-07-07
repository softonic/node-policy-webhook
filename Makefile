BIN := node-policy-webhook
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/softonic/node-policy-webhook
VERSION := 0.0.0-dev
ARCH := amd64
APP := node-policy-webhook
NAMESPACE := default

IMAGE := $(BIN)

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
		/bin/sh -c "ARCH=$(ARCH) VERSION=$(VERSION) PKG=$(PKG) ./build/build"

.PHONY: image
image: build
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

.PHONY: cert
cert:
	ssl/ssl.sh $(APP) $(NAMESPACE)

.PHONY: apply-patch
apply-patch:
	ssl/patch_ca_bundle.sh

.PHONY: clean
clean:
	rm -fr bin .go

.PHONY: undeploy
undeploy:
	kubectl delete -f manifests/ || true

.PHONY: deploy
deploy:
	kubectl create -f manifests/deployment.yaml
	kubectl create -f manifests/service.yaml
	kubectl create -f manifests/mutatingwebhook.yaml

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

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
