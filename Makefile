.PHONY: help dev build test lint docker-up docker-down \
        helm-setup helm-build helm-install helm-upgrade helm-uninstall helm-status helm-open helm-logs helm-teardown

-include .env
export

KIND_CLUSTER  ?= lil-poker
K8S_NAMESPACE ?= lil-poker
HELM_RELEASE  ?= lil-poker
HELM_CHART    := .infra/helm

comma := ,
escaped_comma := \,

FRONTEND_PORT ?= 8090
SMALL_BLIND ?= 10
BIG_BLIND ?= 20
TURN_TIMEOUT_SECS ?= 20
COOKIE_SECRET ?= change-me-in-production
ALLOWED_ORIGINS ?= http://localhost:8090
GATEWAY_HOSTNAME ?= lil.poker.localhost

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:
	cd backend && go build -o ../bin/lil-poker ./cmd/server

test:
	cd backend && go test ./...

lint:
	cd backend && golangci-lint run ./...

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-api-logs:
	docker compose logs -f api

dev-backend:
	cd backend && go run ./cmd/server

dev:
	@$(MAKE) -j2 dev-backend frontend-dev

helm-setup:
	@which kind > /dev/null || (echo "Installing kind..." && \
	  curl -Lo /tmp/kind https://kind.sigs.k8s.io/dl/v0.29.0/kind-linux-amd64 && \
	  chmod +x /tmp/kind && sudo mv /tmp/kind /usr/local/bin/kind)
	@kind get clusters | grep -q $(KIND_CLUSTER) || kind create cluster --name $(KIND_CLUSTER)
	@echo "Installing Gateway API CRDs + Envoy Gateway..."
	@kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/latest/download/install.yaml
	@kubectl wait --namespace envoy-gateway-system \
	  --for=condition=available deployment/envoy-gateway \
	  --timeout=120s
	@echo "Cluster ready."

helm-build:
	docker build -t lil-poker-api:local ./backend
	docker build -t lil-poker-frontend:local ./frontend
	kind load docker-image lil-poker-api:local --name $(KIND_CLUSTER)
	kind load docker-image lil-poker-frontend:local --name $(KIND_CLUSTER)

helm-install:
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(K8S_NAMESPACE) \
	  --create-namespace \
	  --set game.smallBlind=$(SMALL_BLIND) \
	  --set game.bigBlind=$(BIG_BLIND) \
	  --set game.turnTimeoutSecs=$(TURN_TIMEOUT_SECS) \
	  --set auth.cookieSecret=$(COOKIE_SECRET) \
	  --set auth.allowedOrigins="$(subst $(comma),$(escaped_comma),$(ALLOWED_ORIGINS))" \
	  --set postgres.user=$(DB_USER) \
	  --set postgres.password=$(DB_PASSWORD) \
	  --set postgres.database=$(DB_NAME) \
	  --set gateway.hostname=$(GATEWAY_HOSTNAME)

helm-upgrade:
	helm upgrade $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(K8S_NAMESPACE) \
	  --set game.smallBlind=$(SMALL_BLIND) \
	  --set game.bigBlind=$(BIG_BLIND) \
	  --set game.turnTimeoutSecs=$(TURN_TIMEOUT_SECS) \
	  --set auth.cookieSecret=$(COOKIE_SECRET) \
	  --set auth.allowedOrigins="$(subst $(comma),$(escaped_comma),$(ALLOWED_ORIGINS))" \
	  --set postgres.user=$(DB_USER) \
	  --set postgres.password=$(DB_PASSWORD) \
	  --set postgres.database=$(DB_NAME) \
	  --set gateway.hostname=$(GATEWAY_HOSTNAME)

helm-uninstall:
	helm uninstall $(HELM_RELEASE) --namespace $(K8S_NAMESPACE)

helm-status:
	kubectl get all -n $(K8S_NAMESPACE)
	@echo ""
	kubectl get gateway,httproute -n $(K8S_NAMESPACE)

helm-open:
	@echo "Opening http://localhost:$(FRONTEND_PORT) ..."
	kubectl port-forward -n envoy-gateway-system svc/$$(kubectl get svc -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=lil-poker-gateway -o jsonpath='{.items[0].metadata.name}') $(FRONTEND_PORT):80

helm-logs:
	kubectl logs -n $(K8S_NAMESPACE) deployment/api -f

helm-teardown:
	kind delete cluster --name $(KIND_CLUSTER)
