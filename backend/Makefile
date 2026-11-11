GOTOOLCHAIN   := $(shell go env GOVERSION)
export GOTOOLCHAIN

GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
GOVULNCHECK   := go run golang.org/x/vuln/cmd/govulncheck@v1.9.0

.PHONY: run fmt vet lint vuln test cover check

run:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; go run ./cmd/server

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

vuln:
	$(GOVULNCHECK) ./...

test:
	go test -race -cover ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

check: vet lint vuln test
