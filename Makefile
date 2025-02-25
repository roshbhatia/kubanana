GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
DOCKER=docker
PROJECT=kubevent
CONTROLLER_BINARY=kubevent-controller
GO111MODULE=on
export GO111MODULE

.PHONY: all build build-controller clean test deps docker-build \
 kind-setup kind-cleanup kind-rebuild kind-test kind-logs \
 chainsaw-test analyze help

help: ## Display this help message
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

all: build ## Build all binaries

build: build-controller ## Build all components

build-controller: ## Build the controller binary
	$(GOBUILD) -o bin/$(CONTROLLER_BINARY) cmd/controller/main.go

clean: ## Clean up build artifacts
	$(GOCLEAN)
	rm -f bin/$(CONTROLLER_BINARY)

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

deps: ## Download dependencies
	$(GOGET) -v ./...

docker-build: ## Build Docker image for controller
	$(DOCKER) build -t $(PROJECT)-controller:latest \
	--build-arg BINARY_NAME=$(CONTROLLER_BINARY) \
	--build-arg ENTRYPOINT=cmd/controller/main.go \
	-f build/Dockerfile .

docker-deps: ## Pull Docker dependencies
	$(DOCKER) pull busybox
	$(DOCKER) pull nginx

generate: ## Generate code
	./hack/update-codegen.sh

install-tools: ## Install required development tools
	$(GOINSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen
	$(GOINSTALL) k8s.io/code-generator/cmd/...

manifests: ## Generate CRD manifests
	controller-gen crd paths="./pkg/apis/..." output:crd:artifacts:config=deploy/crds

fmt: ## Format the code
	$(GOCMD) fmt ./...

lint: ## Run linter
	golangci-lint run ./...

analyze: lint fmt ## Run all code analysis tools
	go vet ./...

chainsaw-test: ## Run chainsaw tests
	chainsaw test --test-dir chainsaw

kind-setup: docker-deps ## Set up kind cluster
	./hack/kind/setup-kind.sh

kind-cleanup: ## Clean up kind cluster
	./hack/kind/cleanup-kind.sh

kind-rebuild: ## Rebuild and redeploy to kind
	./hack/kind/rebuild-and-redeploy.sh

kind-test: ## Run tests in kind
	./hack/kind/test-event-job.sh

kind-logs: ## View controller logs
	./hack/kind/view-logs.sh controller --follow