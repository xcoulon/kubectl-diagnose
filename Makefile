GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}
GO_PATH_BIN=${GOPATH}/bin
BIN_DIR := bin

.PHONY: build
## Build the binary
build:
	@rm -rf $(BIN_DIR) 2>/dev/null || true
	@echo "building the binary in ${GO_PACKAGE_PATH}"
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TAG:=$(shell git tag --contains $(BUILD_COMMIT)))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	@$(Q)CGO_ENABLED=0 \
		go build ${V_FLAG} \
		-ldflags \
		 " -X github.com/xcoulon/kubectl-diagnose/cmd.BuildCommit=$(BUILD_COMMIT) \
	       -X github.com/xcoulon/kubectl-diagnose/cmd.BuildTag=$(BUILD_TAG) \
	       -X github.com/xcoulon/kubectl-diagnose/cmd.BuildTime=$(BUILD_TIME)" \
		-o $(BIN_DIR)/kubectl-diagnose \
		main.go

.PHONY: test
## run all tests excluding fixtures and vendored packages
test: 
	@go test github.com/xcoulon/kubectl-diagnose/... -v -failfast

.PHONY: install
## installs the binary executable in the $GOPATH/bin directory
install: build
	@cp ${BIN_DIR}/kubectl-diagnose ${GO_PATH_BIN}/

.PHONY: install-golangci-lint
## Install development tools.
install-golangci-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

.PHONY: lint
## run golangci-lint against project
lint: install-golangci-lint
	@$(shell go env GOPATH)/bin/golangci-lint run -v -c .golangci.yml ./...

