SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GOLANGCI_VERSION = 1.24.0

GO ?= go

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

# Install from source.
install:
	@echo "==> Installing up ${GOPATH}/bin/kinesis-tail"
	@$(GO) install ./...
.PHONY: install

# Run all the tests
test:
	@echo "==> Testing"
	@go test -v -covermode=count -coverprofile=coverage.txt ./pkg/... ./cmd/...
.PHONY: test

# Run all the linters
lint: bin/golangci-lint
	@echo "==> Linting"
	@bin/golangci-lint run
.PHONY: lint

# Run all the tests and code checks
ci: test lint
.PHONY: ci

# Release binaries to GitHub.
release:
	@echo "==> Releasing"
	@goreleaser -p 1 --rm-dist -config .goreleaser.yml
	@echo "==> Complete"
.PHONY: release

generate-mocks:
	mockery -dir ../../aws/aws-sdk-go/service/lambda/lambdaiface --all
.PHONY: generate-mocks
