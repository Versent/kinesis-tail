SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GO ?= go

# Install all the build and lint dependencies
setup:
	@$(GO) get -u github.com/alecthomas/gometalinter
	@$(GO) get -u github.com/golang/dep/cmd/dep
	@$(GO) get -u github.com/axw/gocov/...
	@$(GO) get github.com/vektra/mockery/...
	@gometalinter --install
	@dep ensure
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
	gometalinter --deadline 300s --vendor ./...
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
