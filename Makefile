golangci-lint := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

lint:
	$(golangci-lint) run
