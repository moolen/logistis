GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

.PHONY: test
test:
	go test ./...

.PHONY: build-cli
build-cli:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/kubectl-blame ./cmd/kubectl-blame

.PHONY: build-webhook
build-webhook:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/admission-webhook ./cmd/webhook

.PHONY: docker-build
docker-build:
	docker build -t logistis:latest .

.PHONY: cluster
cluster:
	kind create cluster

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster

.PHONY: push
push: docker-build
	kind load docker-image --name external-secrets logistis:latest

.PHONY: deploy-config
deploy-config:
	kubectl apply -f dev/manifests/cluster-config/

.PHONY: delete-config
delete-config:
	kubectl delete -f dev/manifests/cluster-config/

.PHONY: deploy
deploy: push delete deploy-config
	kubectl apply -f dev/manifests/webhook/

.PHONY: delete
delete:
	kubectl delete -f dev/manifests/webhook/ || true

.PHONY: example
example:
	kubectl apply -f dev/manifests/pods/deploy.yaml

.PHONY: delete-example
delete-example:
	kubectl delete -f dev/manifests/pods/deploy.yaml --force
