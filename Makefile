EXAMPLE_FIXTURES := $(sort $(shell find examples -mindepth 2 -maxdepth 2 -name mapture.yaml -exec dirname {} \; | xargs -n1 basename))
FIXTURES := $(EXAMPLE_FIXTURES) playground
FIXTURE ?= demo
PRIMARY_GOAL := $(firstword $(MAKECMDGOALS))
SECOND_GOAL := $(word 2,$(MAKECMDGOALS))
POSITIONAL_FIXTURE_COMMANDS := validate scan graph serve

ifneq ($(filter $(PRIMARY_GOAL),$(POSITIONAL_FIXTURE_COMMANDS)),)
ifneq ($(filter $(SECOND_GOAL),$(FIXTURES)),)
override FIXTURE := $(SECOND_GOAL)
.PHONY: $(SECOND_GOAL)
$(SECOND_GOAL):
	@:
endif
endif

.PHONY: help fixtures build web install-dev-tools install-git-hooks init-hooks \
	test-go test lint vet fmt cli-help \
	testing-help testing-build testing-init playground-init audit-public \
	validate scan graph serve run

help: ## Show grouped development, verification, and fixture commands
	@printf '\n%s\n' "Repo Development Commands"
	@printf '  \033[36m%-18s\033[0m %s\n' "install-dev-tools" "Install local Go dev tools into testing/tools/bin"
	@printf '  \033[36m%-18s\033[0m %s\n' "install-git-hooks" "Configure git to use the repo-managed hooks"
	@printf '  \033[36m%-18s\033[0m %s\n' "init-hooks" "Configure git to use the repo-managed hooks"
	@printf '  \033[36m%-18s\033[0m %s\n' "build" "Build the local mapture binary into build/"
	@printf '  \033[36m%-18s\033[0m %s\n' "web" "Rebuild the frontend bundle under src/internal/webui/dist/"
	@printf '\n%s\n' "Repo Verification Commands"
	@printf '  \033[36m%-18s\033[0m %s\n' "test-go" "Run Go tests through gotestsum"
	@printf '  \033[36m%-18s\033[0m %s\n' "test" "Run the full local verification suite"
	@printf '  \033[36m%-18s\033[0m %s\n' "lint" "Run golangci-lint against src/"
	@printf '  \033[36m%-18s\033[0m %s\n' "vet" "Run go vet against src/"
	@printf '  \033[36m%-18s\033[0m %s\n' "fmt" "Format Go source files under src/"
	@printf '  \033[36m%-18s\033[0m %s\n' "audit-public" "Run public-release hygiene checks against tracked files"
	@printf '  \033[36m%-18s\033[0m %s\n' "cli-help" "Show CLI help from the current source tree"
	@printf '\n%s\n' "Local Verification With Fixtures"
	@printf '  \033[36m%-18s\033[0m %s\n' "fixtures" "List discovered fixtures"
	@printf '  \033[36m%-18s\033[0m %s\n' "testing-help" "Show the testing-first wrapper commands and fixture paths"
	@printf '  \033[36m%-18s\033[0m %s\n' "testing-build" "Build the current source into testing/bin/mapture"
	@printf '  \033[36m%-18s\033[0m %s\n' "testing-init" "Run init against testing/playground"
	@printf '  \033[36m%-18s\033[0m %s\n' "playground-init" "Run init against the gitignored testing playground"
	@printf '  \033[36m%-18s\033[0m %s\n' "validate" "Validate a fixture: make validate FIXTURE=<fixture|all>"
	@printf '  \033[36m%-18s\033[0m %s\n' "scan" "Scan a fixture: make scan FIXTURE=<fixture|all>"
	@printf '  \033[36m%-18s\033[0m %s\n' "graph" "Export Mermaid for a fixture: make graph FIXTURE=<fixture|all>"
	@printf '  \033[36m%-18s\033[0m %s\n' "serve" "Run the local server against a fixture: make serve FIXTURE=<fixture>"
	@printf '  \033[36m%-18s\033[0m %s\n' "run" "Run any CLI command for a fixture: make run FIXTURE=<fixture> CMD=<cli-command>"
	@printf '\n%s\n' "Fixtures"
	@for fixture in $(FIXTURES); do printf '  %s\n' "$$fixture"; done
	@printf '\n%s\n' "Fixture Aliases"
	@for fixture in $(FIXTURES); do \
		printf '  validate.%s  scan.%s  graph.%s  serve.%s\n' "$$fixture" "$$fixture" "$$fixture" "$$fixture"; \
	done

fixtures: ## List discovered fixtures
	@for fixture in $(FIXTURES); do echo "$$fixture"; done

build: ## Build the local mapture binary into build/
	@./scripts/build.sh

web: ## Rebuild the frontend bundle under src/internal/webui/dist/
	@go run ./scripts/build-web

install-dev-tools: ## Install local Go dev tools into testing/tools/bin
	@./scripts/test-go.sh --install-only

install-git-hooks: ## Configure git to use the repo-managed hooks
	@./scripts/install-git-hooks.sh

init-hooks: ## Configure git to use the repo-managed hooks
	@./scripts/init-hooks.sh

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

audit-public: ## Run public-release hygiene checks against tracked files
	@./scripts/audit-public.sh

cli-help: ## Show CLI help from the current source tree
	@go run src/main.go --help

testing-help: ## Show the testing-first wrapper commands and fixture paths
	@./scripts/go.sh help

testing-build: ## Build the current source into testing/bin/mapture
	@./scripts/go.sh build

testing-init: ## Run init against testing/playground
	@./scripts/go.sh init

playground-init: ## Run init against the gitignored testing playground
	@$(MAKE) --no-print-directory testing-init

validate: ## Validate a fixture through testing/: make validate FIXTURE=<fixture|all>
	@./scripts/go.sh validate "$(FIXTURE)"

scan: ## Scan a fixture and write testing/outputs/<fixture>.scan.json; FIXTURE=all scans all examples
	@./scripts/go.sh scan "$(FIXTURE)"

graph: ## Export Mermaid for a fixture into testing/outputs/<fixture>.mmd; FIXTURE=all graphs all examples
	@./scripts/go.sh graph "$(FIXTURE)"

serve: ## Rebuild testing/bin/mapture and run the local server against a fixture
	@./scripts/go.sh serve "$(FIXTURE)"

run: ## Run any CLI command for a fixture: make run FIXTURE=<fixture> CMD=<cli-command>
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run FIXTURE=<fixture> CMD="<cli-command>"'; exit 1; fi
	@./scripts/go.sh fixture "$(FIXTURE)" $(CMD)

define MAKE_FIXTURE_TARGETS
.PHONY: validate.$(1) scan.$(1) graph.$(1) serve.$(1)

validate.$(1):
	@$(MAKE) --no-print-directory validate FIXTURE=$(1)

scan.$(1):
	@$(MAKE) --no-print-directory scan FIXTURE=$(1)

graph.$(1):
	@$(MAKE) --no-print-directory graph FIXTURE=$(1)

serve.$(1):
	@$(MAKE) --no-print-directory serve FIXTURE=$(1)
endef

$(foreach fixture,$(FIXTURES),$(eval $(call MAKE_FIXTURE_TARGETS,$(fixture))))
