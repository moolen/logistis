GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# check if there are any existing `git tag` values
ifeq ($(shell git tag),)
# no tags found - default to initial tag `v0.0.0`
export VERSION := $(shell echo "v0.0.0-$$(git rev-list HEAD --count)-g$$(git describe --dirty --always)" | sed 's/-/./2' | sed 's/-/./2')
else
# use tags
export VERSION := $(shell git describe --dirty --always --tags --exclude 'helm*' | sed 's/-/./2' | sed 's/-/./2')
endif

DOCKER_BUILD_ARGS ?=
DOCKERFILE := Dockerfile
IMAGE_REPO := logistis
IMAGE_TAG := $(VERSION)
ARCH := $(shell go env GOARCH)
PLATFORM := linux/$(ARCH)

.PHONY: test
test:
	go test ./...

.PHONY: build-cli
build-cli:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/kubectl-blame ./cmd/kubectl-blame

.PHONY: build-webhook
build-webhook:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/admission-webhook ./cmd/webhook

.PHONY: docker.build
docker.build:
	docker buildx build \
		--platform $(PLATFORM) \
		$(DOCKER_BUILD_ARGS) \
		-t $(IMAGE_REPO):$(IMAGE_TAG) .

.PHONY: docker.push.kind
docker.push.kind:
	kind load docker-image --name kind $(IMAGE_REPO):$(IMAGE_TAG)

# ---- local development
.PHONY: cluster
cluster:
	kind create cluster

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster

.PHONY: gen-tls
gen-tls:
	./dev/gen-certs.sh logistis.default.svc

.PHONY: deploy
deploy: gen-tls docker.build docker.push.kind
	helm upgrade --install logistis ./chart/logistis \
		--values ./dev/values.dev.yaml \
		--wait

# -------
# EXAMPLES
.PHONY: example
example:
	kubectl apply -f dev/deploy.yaml

.PHONY: delete-example
delete-example:
	kubectl delete -f dev/deploy.yaml --force
