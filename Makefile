.PHONY: build run docker-build cluster deploy all clean

CHART_DIR := deploy/repository-tool
RELEASE_NAME := repository-tool
NAMESPACE := repository-tool
IMAGE := repository-tool
CLUSTER := repository-tool
PORT ?= 8080
PORT_FILE := bin/.cluster-port

# Shell snippet that sets shell var $$p to the first free TCP port at/after $(PORT).
# Inlined into recipes (rather than a recursive make call) so the captured value is
# only the loop's own output and not contaminated by sub-make tracing/diagnostics.
define FIND_FREE_PORT
p=$(PORT); n=0; \
	while lsof -iTCP:$$p -sTCP:LISTEN -Pn >/dev/null 2>&1; do \
		n=$$((n+1)); [ $$n -ge 50 ] && { echo "no free port found near $(PORT)" >&2; exit 1; }; \
		p=$$((p+1)); \
	done; \
	if [ "$$p" != "$(PORT)" ]; then echo ">> port $(PORT) busy, using $$p"; fi
endef

# Build the Go binary
build:
	go build -o bin/$(IMAGE) ./cmd/server

# Run locally (requires GITHUB_TOKEN env var). Auto-picks a free port if $(PORT) is busy.
run:
	@$(FIND_FREE_PORT); \
		PORT=$$p go run ./cmd/server

# Build Docker image
docker-build:
	docker build -t $(IMAGE):latest .

# Create kind cluster with port mapping. Auto-picks a free host port if $(PORT) is busy.
cluster:
	@mkdir -p bin
	@$(FIND_FREE_PORT); \
		echo $$p > $(PORT_FILE); \
		sed "s/hostPort: 8080/hostPort: $$p/" deploy/kind-config.yaml | \
			kind create cluster --name $(CLUSTER) --config=-

# Load Docker image into kind cluster
load: docker-build
	kind load docker-image $(IMAGE):latest --name $(CLUSTER)

# Deploy to kind cluster via Helm (set your GitHub token!)
deploy: load
	helm upgrade --install $(RELEASE_NAME) $(CHART_DIR) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set github.token=$(GITHUB_TOKEN)

# Full setup: create cluster + build + deploy
all: cluster deploy
	@p=$$(cat $(PORT_FILE) 2>/dev/null || echo $(PORT)); \
		echo ""; \
		echo "$(RELEASE_NAME) is deploying. Check status with:"; \
		echo "  kubectl -n $(NAMESPACE) get pods"; \
		echo ""; \
		echo "Once running, access at: http://localhost:$$p"

# Render Helm templates without installing
template:
	helm template $(RELEASE_NAME) $(CHART_DIR) --namespace $(NAMESPACE)

# Lint the Helm chart
lint:
	helm lint $(CHART_DIR)

# Check deployment status
status:
	kubectl -n $(NAMESPACE) get pods
	kubectl -n $(NAMESPACE) get svc

# View logs
logs:
	kubectl -n $(NAMESPACE) logs -l app.kubernetes.io/name=$(IMAGE) -f

# Uninstall the Helm release
uninstall:
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)

# Clean up everything
clean:
	-helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE) 2>/dev/null
	kind delete cluster --name $(CLUSTER)
	rm -f bin/$(IMAGE) $(PORT_FILE)

# Run locally with Docker. Auto-picks a free host port if $(PORT) is busy.
docker-run:
	@$(FIND_FREE_PORT); \
		docker run -p $$p:8080 -e GITHUB_TOKEN=$(GITHUB_TOKEN) $(IMAGE):latest
