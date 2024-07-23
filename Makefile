golangci-lint := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
gofmtmd := go run github.com/po3rin/gofmtmd/cmd/gofmtmd@latest

lint:
	$(golangci-lint) run

fmtmd:
	$(gofmtmd) -r README.md
