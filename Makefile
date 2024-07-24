golangci-lint := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
gofmtmd := go run github.com/po3rin/gofmtmd/cmd/gofmtmd@latest

lint:
	$(golangci-lint) run

test:
	go test ./... -shuffle=on -race

test.cover:
	go test -race -shuffle=on -coverprofile=coverage.txt -covermode=atomic ./...

# For local environment
cov:
	go test -cover -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

fmtmd:
	$(gofmtmd) -r README.md
