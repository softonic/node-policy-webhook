BIN := node-policy-webhook
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/softonic/node-policy-webhook
VERSION := 0.0.0-dev
ARCH := amd64
APP := node-policy-webhook
NAMESPACE := default
KO_DOCKER_REPO = registry.softonic.io/node-policy-webhook

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.14-buster


deploy-prod: export IMAGE_GEN = "github.com/softonic/node-policy-webhook/cmd/node-policy-webhook"

deploy-dev:  export IMAGE_GEN = $(APP):$(VERSION)


.PHONY: all
all: dev

.PHONY: start
start: dev deploy-dev

.PHONY: build
build: generate
	go mod download
	GOARCH=${ARCH} go install -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./cmd/node-policy-webhook/.../

.PHONY: test
test:
	GOARCH=${ARCH} go test -v -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./...

.PHONY: image
image:
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

.PHONY: dev
dev: image
	kind load docker-image $(IMAGE):$(VERSION)

.PHONY: cert
cert:
	ssl/ssl.sh $(APP) $(NAMESPACE)

.PHONY: apply-patch
apply-patch: cert
	ssl/patch_ca_bundle.sh

.PHONY: undeploy
undeploy:
	kubectl delete -f manifests/ || true

.PHONY: deploy-dev
deploy: apply-patch
	cat manifests/deployment-tpl.yaml | envsubst > manifests/deployment.yaml
	kubectl apply -f manifests/noodepolicies.softonic.io_nodepolicyprofiles.yaml
	kubectl apply -f manifests/deployment.yaml
	kubectl delete pod $$(kubectl get pods --selector=app=node-policy-webhook -o jsonpath='{.items..metadata.name}')
	kubectl apply -f manifests/service.yaml
	kubectl apply -f manifests/mutatingwebhook.yaml
	kubectl apply -f manifests/nodepolicyprofile_viewer_role.yaml
	kubectl apply -f manifests/role_binding.yaml
	kubectl apply -f samples/nodepolicyprofile.yaml

.PHONY: up
up: image undeploy deploy

.PHONY: docker-push
docker-push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=manifests

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
