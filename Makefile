# ====================
# Formatting & linting
# ====================
# Lint checking consists of two parts: "golangci-lint" and "gci". Both should be used together.
# "gci" may still flag issues even if the golangci-lint check passes.
# Therefore, if any files need formatting, they will be printed, and the final exit code will be
# successful only if no such files exist.
.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml && (gci list . | sed 's/^/BadFormat: /'; [ $$(gci list . | wc -c) -eq 0 ])

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
