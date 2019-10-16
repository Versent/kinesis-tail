SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GO ?= go

# Install all the build and lint dependencies
setup:
ifeq ( $(wildcard $(go env GOPATH)/bin/golangci-lint), )
	@curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| \
		sh -s -- -b $(go env GOPATH)/bin v1.20.0
endif
	@$(GO) get -u github.com/axw/gocov/...
	@$(GO) get github.com/vektra/mockery/...
	@go mod vendor
.PHONY: setup

# Install from source.
install:
	@echo "==> Installing up ${GOPATH}/bin/kinesis-tail"
	@$(GO) install ./...
.PHONY: install

# Run all the tests
test:
	@gocov test -timeout=2m ./... | gocov report
.PHONY: test

# Run all the linters
lint:
	$(GOPATH)/bin/golangci-lint run
.PHONY: lint

# Run all the tests and code checks
ci: setup test lint
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
