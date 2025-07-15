ARGOCD_VERSION ?= v3.0.11
KIND_CONFIG ?= kind-config.yaml
KIND_CLUSTER_NAME ?= `yq .name -r ${KIND_CONFIG}`
TAG ?= 0.1.0
REPOSITORY ?= ghcr.io/estenrye

.PHONY: build-local kind-create kind-delete

build-local:
	docker buildx build -t $(REPOSITORY)/argocd-namespace-generator:$(TAG) .

kind-create: build-local
	kind create cluster --config $(KIND_CONFIG)
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl create namespace argocd
	kind load docker-image $(REPOSITORY)/argocd-namespace-generator:$(TAG) --name $(KIND_CLUSTER_NAME)
	kubectl apply -k examples/manifests

kind-delete:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kind delete cluster --name $(KIND_CLUSTER_NAME)

logs-applicationsetcontroller:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl logs -n argocd -l app.kubernetes.io/name=argocd-applicationset-controller -f

logs-applicationcontroller:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller -f

logs-plugin:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl logs -n argocd -l app=argocd-namespace-generator -f

describe-whoami-appset:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl describe applicationset -n argocd whoami

delete-whoami-appset:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl delete -k ./examples/applications

apply-whoami-appset:
	kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	kubectl apply -k ./examples/applications