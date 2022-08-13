GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/admission-webhook ./cmd/webhook

.PHONY: docker-build
docker-build:
	docker build -t simple-kubernetes-webhook:latest .

.PHONY: cluster
cluster:
	kind create cluster

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster

.PHONY: push
push: docker-build
	kind load docker-image --name external-secrets simple-kubernetes-webhook:latest

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

.PHONY: pod
pod:
	kubectl apply -f dev/manifests/pods/lifespan-seven.pod.yaml

.PHONY: delete-pod
delete-pod:
	kubectl delete -f dev/manifests/pods/lifespan-seven.pod.yaml --force
