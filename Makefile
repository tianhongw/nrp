ROOT_DIR = $(CURDIR)
BUILD_DIR = build
NRPS_DIR = cmd/nrps
NRPC_DIR = cmd/nrpc
NRPS_NAME = nrps
NRPC_NAME = nrpc

CURR_TIME := $(shell date +"%a %b %d %Y %H:%M:%S GMT%z")

GIT_SUMMARY := $(shell git describe --tags --dirty --always)
GIT_BRANCH := $(shell git symbolic-ref -q --short HEAD || echo "none")
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_COMMIT_TIME := $(shell git log -1 --format=%cd --date=format:'%a %b %d %Y %H:%M:%S GMT%z')

LDFLAGS = -X 'github.com/tianhongw/nrp/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/tianhongw/nrp/version.GitBranch=$(GIT_BRANCH)' \
	-X 'github.com/tianhongw/nrp/version.GitSummary=$(GIT_SUMMARY)' \
	-X 'github.com/tianhongw/nrp/version.GitCommitTime=$(GIT_COMMIT_TIME)' \
	-X 'github.com/tianhongw/nrp/version.BuildTime=$(CURR_TIME)'

PACKAGES := $(shell go list ./...)

define fail
	@echo "$(CURR_NAME): $(1)" >&2
	@exit 1
endef

.PHONY: build

build_server:
	# build for linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(NRPS_NAME) $(NRPS_DIR)/main.go

build_client:
	# build for linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(NRPC_NAME) $(NRPC_DIR)/main.go

build_all: build_server build_client

build_server_osx:
	# build for darwin amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(NRPS_NAME) $(NRPS_DIR)/main.go

build_client_osx:
	# build for darwin amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(NRPC_NAME) $(NRPC_DIR)/main.go

build_osx_all: build_server_osx build_client_osx

clean:
	@go clean
	@test -d $(BUILD_DIR) && rm -rf $(BUILD_DIR) || true
	@test -f $(NRPS_NAME).tar.gz && rm -f $(NRPS_NAME).tar.gz || true

check:
	@go vet ./...

test:
	@go test ./...

fmt:
	for pkg in ${PACKAGES}; do \
		go fmt $$pkg; \
	done;
