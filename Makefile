lint:
	golangci-lint run -c .golangci.yml --out-format colored-line-number

fix:
	golangci-lint run -c .golangci.yml --out-format colored-line-number --fix

format:
	gofumpt -w .
