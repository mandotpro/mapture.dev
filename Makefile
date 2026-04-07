.PHONY: help build test test-go vet cli-help validate-demo validate-ecommerce install-dev-tools install-git-hooks playground-init run-demo run-playground fmt

help: ## Show available development commands
	@awk 'BEGIN {FS = ": ## "}; /^[a-zA-Z0-9_.-]+: ## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the local mapture binary into build/
	@./scripts/build.sh

test-go: ## Run Go tests through gotestsum
	@./scripts/test-go.sh

test: ## Run the full local verification suite
	@./scripts/test.sh

vet: ## Run go vet against src/
	@go vet ./src/...

fmt: ## Format Go source files under src/
	@find ./src -name '*.go' -print0 | xargs -0 gofmt -w

cli-help: ## Show CLI help from the current source tree
	@go run src/main.go --help

validate-demo: ## Validate the canonical demo fixture
	@./scripts/go.sh validate-demo

validate-ecommerce: ## Validate the polyglot ecommerce fixture
	@go run src/main.go validate examples/ecommerce

install-dev-tools: ## Install local Go dev tools into testing/tools/bin
	@./scripts/test-go.sh --install-only

install-git-hooks: ## Configure git to use the repo-managed hooks
	@./scripts/install-git-hooks.sh

playground-init: ## Run init against the gitignored testing playground
	@./scripts/go.sh init

run-demo: ## Run the CLI against examples/demo: make run-demo CMD="validate"
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run-demo CMD="validate"'; exit 1; fi
	@./scripts/go.sh demo $(CMD)

run-playground: ## Run the CLI against testing/playground: make run-playground CMD="validate"
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run-playground CMD="validate"'; exit 1; fi
	@./scripts/go.sh playground $(CMD)
