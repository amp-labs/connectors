# ====================
# Formatting & linting
# ====================
# Lint checking consists of two parts: "golangci-lint" and "gci". Both should be used together.
# "gci" may still flag issues even if the golangci-lint check passes.
# Therefore, if any files need formatting, they will be printed, and the final exit code will be
# successful only if no such files exist.
.PHONY: lint
lint: custom-gcl
	@output="$$(./custom-gcl run -c .golangci.yml 2>&1)"; \
	echo "$$output"; \
	if echo "$$output" | grep -Eq "build linters|module.* not found"; then \
		echo "❌ GolangCI-Lint plugin build failed. Try 'make linter-rebuild'."; \
		exit 1; \
	fi; \
	gci list . | sed 's/^/BadFormat: /'; \
	[ $$(gci list . | wc -c) -eq 0 ]


# Build custom golangci-lint binary with linter plugins nogoroutine, modulelinter.
.PHONY: custom-gcl
custom-gcl:
	@if [ ! -f custom-gcl ]; then \
		echo "Building custom golangci-lint binary with nogoroutine & module linter..."; \
		golangci-lint custom; \
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
# For ones that do not have auto-fix enabled by golangci-lint (e.g. wsl and gci), we add the fix commands manually to this list.
.PHONY: fix
fix: custom-gcl
	wsl --allow-cuddle-declarations --fix ./... && \
		gci write . && \
		./custom-gcl run -c .golangci.yml --fix

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

# Set up git hooks which will use PR templates on branch push.
.PHONY: git-hooks-install
git-hooks-install:
	. ./scripts/git/install_hooks.sh; install

.PHONY: git-hooks-uninstall
git-hooks-uninstall:
	. ./scripts/git/install_hooks.sh; uninstall

# Compiles connector generator CLI. For more information see scripts/connectorgen/README.md
.PHONY: connector-gen
connector-gen:
	go build -o ./bin/cgen ./scripts/connectorgen/main.go && echo "now run command: ./bin/cgen"
