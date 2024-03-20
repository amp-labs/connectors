# ====================
# Formatting & linting
# ====================
.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

# Run a few autoformatters and print out unfixable errors
# PRE-REQUISITES: install linters, see https://ampersand.slab.com/posts/engineering-onboarding-guide-environment-set-up-9v73t3l8#huik9-install-linters
# If you're curious, run `golangci-lint help linters` to see which linters have auto-fix enabled by golangci-lint.
# For ones that do not have auto-fix enabled by golangci-lint (e.g. wsl and gci), we add the fix commands manually to this list.
.PHONY: fix
fix:
	wsl --allow-cuddle-declarations --fix ./... && \
		gci write . && \
		golangci-lint run -c .golangci.yml --fix

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort

# Alias for fix
.PHONY: format
format: fix

.PHONY: test
test:
	go test -v ./...
