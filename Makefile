SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GO ?= go

# Install all the build and lint dependencies
setup:
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/pierrre/gotestcover
	go get -u golang.org/x/tools/cmd/cover
	go get github.com/vektra/mockery/...
	gometalinter --install
.PHONY: setup

# Install from source.
install:
	@echo "==> Installing up ${GOPATH}/bin/kinesis-tail"
	@$(GO) install ./...
.PHONY: install

# Run all the tests
test:
	@gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	@$(GO) tool cover -html=coverage.txt
.PHONY: cover

# Run all the linters
lint:
	gometalinter --vendor ./...
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

.PHONY: setup test cover generate-mocks