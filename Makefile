EXAMPLE_FIXTURES := $(sort $(shell find examples -mindepth 2 -maxdepth 2 -name mapture.yaml -exec dirname {} \; | xargs -n1 basename))
FIXTURES := $(EXAMPLE_FIXTURES) playground
FIXTURE ?= demo
PRIMARY_GOAL := $(firstword $(MAKECMDGOALS))
SECOND_GOAL := $(word 2,$(MAKECMDGOALS))
POSITIONAL_FIXTURE_COMMANDS := validate scan export-json-graph export-json-visualisation serve

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
	validate scan export-json-graph export-json-visualisation serve run

help: ## Show grouped development, verification, and fixture commands
	@./scripts/help.sh

fixtures: ## List discovered fixtures
	@for fixture in $(FIXTURES); do echo "$$fixture"; done

build: ## Build the latest local binary into build/, refreshing embedded web if stale
	@./scripts/build.sh

web: ## Rebuild only the frontend bundle under src/internal/webui/dist/
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

validate: ## Validate a fixture through testing/: make validate ecommerce or FIXTURE=ecommerce
	@./scripts/go.sh validate "$(FIXTURE)"

scan: ## Scan a fixture and write testing/outputs/<fixture>.scan.json; FIXTURE=all scans all examples
	@./scripts/go.sh scan "$(FIXTURE)"

export-json-graph: ## Export JGF for a fixture into testing/outputs/<fixture>.graph.json; FIXTURE=all exports all examples
	@./scripts/go.sh export-json-graph "$(FIXTURE)"

export-json-visualisation: ## Export explorer JSON into testing/outputs/<fixture>.visualisation.json; FIXTURE=all exports all examples
	@./scripts/go.sh export-json-visualisation "$(FIXTURE)"

serve: ## Always rebuild the embedded UI and binary, then run the local server against a fixture
	@./scripts/go.sh serve "$(FIXTURE)"

run: ## Run any CLI command for a fixture: make run FIXTURE=<fixture> CMD=<cli-command>
	@if [ -z "$(CMD)" ]; then echo 'Usage: make run FIXTURE=<fixture> CMD="<cli-command>"'; exit 1; fi
	@./scripts/go.sh fixture "$(FIXTURE)" $(CMD)

define MAKE_FIXTURE_TARGETS
.PHONY: validate.$(1) scan.$(1) export-json-graph.$(1) export-json-visualisation.$(1) serve.$(1)

validate.$(1):
	@$(MAKE) --no-print-directory validate FIXTURE=$(1)

scan.$(1):
	@$(MAKE) --no-print-directory scan FIXTURE=$(1)

export-json-graph.$(1):
	@$(MAKE) --no-print-directory export-json-graph FIXTURE=$(1)

export-json-visualisation.$(1):
	@$(MAKE) --no-print-directory export-json-visualisation FIXTURE=$(1)

serve.$(1):
	@$(MAKE) --no-print-directory serve FIXTURE=$(1)
endef

$(foreach fixture,$(FIXTURES),$(eval $(call MAKE_FIXTURE_TARGETS,$(fixture))))
