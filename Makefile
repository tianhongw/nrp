ROOT_DIR = $(CURDIR)
BUILD_DIR = build
CURR_NAME = nrp

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

build:
	# build for linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(CURR_NAME) main.go

build_osx:
	# build for darwin amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -gcflags "all=-N -l" -v -o $(BUILD_DIR)/$(CURR_NAME)_darwin_amd64 main.go

clean:
	@go clean
	@test -d $(BUILD_DIR) && rm -rf $(BUILD_DIR) || true
	@test -f $(CURR_NAME).tar.gz && rm -f $(CURR_NAME).tar.gz || true

check:
	@go vet ./...

test:
	@go test ./...

fmt:
	for pkg in ${PACKAGES}; do \
		go fmt $$pkg; \
	done;
