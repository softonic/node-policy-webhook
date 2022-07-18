BIN := node-policy-webhook
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/softonic/node-policy-webhook
VERSION ?= 0.1.6-dev
ARCH ?= amd64
APP ?= node-policy-webhook
NAMESPACE ?= default
RELEASE_NAME ?= node-policy-webhook
KO_DOCKER_REPO = registry.softonic.io/node-policy-webhook
REPOSITORY ?= node-policy-webhook

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.14-buster


deploy-prod: export IMAGE_GEN = "github.com/softonic/node-policy-webhook/cmd/node-policy-webhook"

deploy:  export IMAGE_GEN = $(APP):$(VERSION)


.PHONY: all
all: dev

.PHONY: build
build: generate
	go mod download
	GOARCH=${ARCH} go build -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./cmd/node-policy-webhook/.../

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

.PHONY: undeploy
undeploy:
	kubectl delete -f manifest.yaml || true

.PHONY: deploy
deploy: manifest
	kubectl apply -f manifest.yaml

.PHONY: up
up: image undeploy deploy

.PHONY: docker-push
docker-push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: secret-values
secret-values:
	./hack/generate_helm_cert_secrets $(APP) $(NAMESPACE)

.PHONY: manifest
manifest: controller-gen helm-chart secret-values
	docker run --rm -v $(PWD):/app -w /app/ alpine/helm:3.2.3 template --release-name $(RELEASE_NAME) --set "image.tag=$(VERSION)" --set "image.repository=$(REPOSITORY)"  -f chart/node-policy-webhook/values.yaml -f chart/node-policy-webhook/secret.values.yaml chart/node-policy-webhook > manifest.yaml

.PHONY: helm-chart
helm-chart: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) webhook paths="./..." output:crd:artifacts:config=chart/node-policy-webhook/templates

.PHONY: helm-deploy
helm-deploy: helm-chart secret-values
	helm upgrade --install $(RELEASE_NAME) --namespace $(NAMESPACE) --set "image.tag=$(VERSION)" -f chart/node-policy-webhook/values.yaml -f chart/node-policy-webhook/secret.values.yaml chart/node-policy-webhook

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) crd:crdVersions=v1 object:headerFile="hack/boilerplate.go.txt" paths="./..."

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
	go get -d sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
