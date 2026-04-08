.PHONY: help build test test-go lint vet fmt web cli-help install-dev-tools install-git-hooks init-hooks \
	playground-init testing-help testing-build testing-init \
	testing-demo-validate testing-demo-scan testing-demo-graph testing-demo-web \
	testing-ecommerce-validate testing-ecommerce-scan testing-ecommerce-graph testing-ecommerce-web \
	testing-playground-validate testing-playground-scan testing-playground-graph testing-playground-web \
	validate-demo validate-ecommerce run-demo run-playground

help: ## Show available development commands
	@awk 'BEGIN {FS = ": ## "}; /^[a-zA-Z0-9_.-]+: ## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the local mapture binary into build/
	@./scripts/build.sh

web: ## Rebuild the frontend bundle under src/internal/webui/dist/
	@go run ./scripts/build-web

test-go: ## Run Go tests through gotestsum
	@./scripts/test-go.sh

test: ## Run the full local verification suite
	@./scripts/test.sh

lint: ## Run golangci-lint against src/
	@./scripts/lint-go.sh

vet: ## Run go vet against src/
	@go vet ./src/...

fmt: ## Format Go source files under src/
	@find ./src -name '*.go' -print0 | xargs -0 gofmt -w

cli-help: ## Show CLI help from the current source tree
	@go run src/main.go --help

testing-help: ## Show the testing-first wrapper commands and fixture paths
	@./scripts/go.sh help

testing-build: ## Build the current source into testing/bin/mapture
	@./scripts/go.sh build

testing-init: ## Run init against testing/playground
	@./scripts/go.sh init

testing-demo-validate: ## Validate examples/demo through testing/
	@./scripts/go.sh validate demo

testing-demo-scan: ## Scan examples/demo and write testing/outputs/demo.scan.json
	@./scripts/go.sh scan demo

testing-demo-graph: ## Export Mermaid for examples/demo into testing/outputs/demo.mmd
	@./scripts/go.sh graph demo

testing-demo-web: ## Run the web UI for examples/demo on http://127.0.0.1:8766
	@./scripts/go.sh web demo

testing-ecommerce-validate: ## Validate examples/ecommerce through testing/
	@./scripts/go.sh validate ecommerce

testing-ecommerce-scan: ## Scan examples/ecommerce and write testing/outputs/ecommerce.scan.json
	@./scripts/go.sh scan ecommerce

testing-ecommerce-graph: ## Export Mermaid for examples/ecommerce into testing/outputs/ecommerce.mmd
	@./scripts/go.sh graph ecommerce

testing-ecommerce-web: ## Run the web UI for examples/ecommerce on http://127.0.0.1:8765
	@./scripts/go.sh web ecommerce

testing-playground-validate: ## Validate testing/playground through testing/
	@./scripts/go.sh validate playground

testing-playground-scan: ## Scan testing/playground and write testing/outputs/playground.scan.json
	@./scripts/go.sh scan playground

testing-playground-graph: ## Export Mermaid for testing/playground into testing/outputs/playground.mmd
	@./scripts/go.sh graph playground

testing-playground-web: ## Run the web UI for testing/playground on http://127.0.0.1:8767
	@./scripts/go.sh web playground

validate-demo: ## Validate the canonical demo fixture
	@$(MAKE) --no-print-directory testing-demo-validate

validate-ecommerce: ## Validate the polyglot ecommerce fixture
	@$(MAKE) --no-print-directory testing-ecommerce-validate

install-dev-tools: ## Install local Go dev tools into testing/tools/bin
	@./scripts/test-go.sh --install-only

install-git-hooks: ## Configure git to use the repo-managed hooks
	@./scripts/install-git-hooks.sh

init-hooks: ## Configure git to use the repo-managed hooks
	@./scripts/init-hooks.sh

playground-init: ## Run init against the gitignored testing playground
	@$(MAKE) --no-print-directory testing-init

run-demo: ## Run the CLI against examples/demo: make run-demo CMD="validate"
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run-demo CMD="validate"'; exit 1; fi
	@./scripts/go.sh demo $(CMD)

run-playground: ## Run the CLI against testing/playground: make run-playground CMD="validate"
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run-playground CMD="validate"'; exit 1; fi
	@./scripts/go.sh playground $(CMD)
