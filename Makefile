ARGOCD_VERSION ?= v3.0.11
KIND_CONFIG ?= kind-config.yaml
KIND_CLUSTER_NAME ?= `yq .name -r ${KIND_CONFIG}`
TAG ?= 0.1.0
REPOSITORY ?= ghcr.io/estenrye

.PHONY: build-local kind-create kind-delete

build-local:
	docker buildx build -t $(REPOSITORY)/argocd-namespace-generator:$(TAG) .

kind-create: kind-delete build-local
	kind create cluster --config $(KIND_CONFIG)
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl create namespace argocd
	kind load docker-image $(REPOSITORY)/argocd-namespace-generator:$(TAG) --name $(KIND_CLUSTER_NAME)
	kubectl apply -k examples/manifests
	kubectl apply -k examples/applications

kind-delete:
	kind delete cluster --name $(KIND_CLUSTER_NAME)