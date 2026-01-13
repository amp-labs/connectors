# ====================
# Formatting & linting
# ====================

# Linter versions - keep in sync with .github/workflows/linter.yml
GOLANGCI_LINT_VERSION=v2.7.1

# Install all linters required by make fix
.PHONY: install/linters
install/linters:
	@echo "Installing gci..."
	go install github.com/daixiang0/gci@latest
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "Installing typos..."
	cargo install typos-cli
	@echo "Linters installed successfully!"

# Install git hooks for this repository
.PHONY: install/hooks
install/hooks:
	@git config --local core.hooksPath scripts/git-hooks
	@echo "Git hooks configured successfully!"

# Install linters and git hooks
.PHONY: install/dev
install/dev: install/linters install/hooks
	@echo "Development environment setup complete!"

# Check linter versions and auto-install if needed
.PHONY: check-linters
check-linters:
	@if ! command -v gci >/dev/null 2>&1; then \
		echo "gci not found, installing..."; \
		go install github.com/daixiang0/gci@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found, installing $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	else \
		INSTALLED_VERSION=$$(golangci-lint --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
		if [ "$$INSTALLED_VERSION" != "$(GOLANGCI_LINT_VERSION)" ]; then \
			echo "golangci-lint version mismatch: installed $$INSTALLED_VERSION, expected $(GOLANGCI_LINT_VERSION)"; \
			echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
		fi \
	fi
	@if ! command -v typos >/dev/null 2>&1; then \
		echo "typos not found, installing..."; \
		cargo install typos-cli; \
	fi

# Build custom golangci-lint binary with linter plugins nogoroutine, modulelinter.
.PHONY: custom-gcl
custom-gcl: check-linters
	@if [ ! -f custom-gcl ]; then \
		echo "Building custom golangci-lint binary with nogoroutine & module linter..."; \
		golangci-lint custom --verbose || exit 1; \
	fi

# Builds custom golangci-lint binary printing the details.
.PHONY: linter-rebuild
linter-rebuild:
	golangci-lint custom --verbose

# Invalidates golangci-lint cache.
.PHONY: linter-clear-cache
linter-clear-cache:
	golangci-lint cache clean

# Run a few autoformatters and print out unfixable errors
# PRE-REQUISITES: install linters, see https://ampersand.slab.com/posts/engineering-onboarding-guide-environment-set-up-9v73t3l8#huik9-install-linters
# If you're curious, run `golangci-lint help linters` to see which linters have auto-fix enabled by golangci-lint.
# For ones that do not have auto-fix enabled by golangci-lint (e.g. gci), we add the fix commands manually to this list.
.PHONY: fix
fix: custom-gcl
	gci write --skip-generated . && \
		./custom-gcl run -c .golangci.yml --fix && \
		typos --config .typos.toml --write-changes

.PHONY: lint
lint: fix

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort

# Alias for fix
.PHONY: format
format: fix

.PHONY: test
test:
	go test -v ./...

.PHONY: test-parallel
test-parallel:
	go test -v ./... -parallel=8 -count=3

.PHONY: test-pretty
test-pretty:
	go run gotest.tools/gotestsum@latest

# Creates PR URLs for each template
# Click on one of them or manually add ?template=<file.md> to the URL if you are creating a PR via the Github website
# Templates: Under github/PULL_REQUEST_TEMPLATE directory you can add more templates
.PHONY: pr-template
pr-template:
	. ./scripts/bash/pr_options.sh; pr_template

# Compiles connector generator CLI. For more information see scripts/connectorgen/README.md
.PHONY: connector-gen
connector-gen:
	go build -o ./bin/cgen ./scripts/connectorgen/main.go && echo "now run command: ./bin/cgen"
