GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}
GO_PATH_BIN=${GOPATH}/bin
BIN_DIR := bin
GIT_COMMIT_ID := $(shell git rev-parse HEAD)
ifneq ($(shell git status --porcelain --untracked-files=no),)
       GIT_COMMIT_ID := $(GIT_COMMIT_ID)-dirty
endif
GIT_COMMIT_ID_SHORT := $(shell git rev-parse --short HEAD)
ifneq ($(shell git status --porcelain --untracked-files=no),)
       GIT_COMMIT_ID_SHORT := $(GIT_COMMIT_ID_SHORT)-dirty
endif
BUILD_TIME = `date -u '+%Y-%m-%dT%H:%M:%SZ'`

.PHONY: build
## Build the binary
build:
	@rm -rf $(BIN_DIR) 2>/dev/null || true
	@echo "building the binary in ${GO_PACKAGE_PATH}"
	@$(Q)CGO_ENABLED=0 \
		go build ${V_FLAG} \
		-ldflags "-X ${GO_PACKAGE_PATH}/pkg/version.Commit=${GIT_COMMIT_ID} -X ${GO_PACKAGE_PATH}/pkg/version.BuildTime=${BUILD_TIME}" \
		-o $(BIN_DIR)/kubectl-diagnose \
		main.go

.PHONY: install-ginkgo
## Install development tools.
install-ginkgo:
	@go install -v github.com/onsi/ginkgo/v2/ginkgo
	@ginkgo version

.PHONY: test
## run all tests excluding fixtures and vendored packages
test: install-ginkgo
	@ginkgo -r --randomize-all --randomize-suites  --trace --race --compilers=0

COVERPKGS := $(shell go list ./... | grep -v vendor | paste -sd "," -)

.PHONY: test-with-coverage
## run all tests excluding fixtures and vendored packages
test-with-coverage: install-ginkgo
	@echo $(COVERPKGS)
	@ginkgo -r --randomize-all --randomize-suites  --trace --race --compilers=0  --cover -coverpkg $(COVERPKGS)

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

