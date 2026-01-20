GOMOD ?= on
GO ?= GO111MODULE=$(GOMOD) go
BINPATH := $(abspath ./bin)

#retrieve go version details for version check
GO_VERSION     := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_VERSION_MAJ := $(shell echo $(GO_VERSION) | cut -f1 -d'.')
GO_VERSION_MIN := $(shell echo $(GO_VERSION) | cut -f2 -d'.')

GOFMT ?= gofmt
RM = rm

GOBINPATH     := $(shell $(GO) env GOPATH)/bin
COMMIT        := $(shell git rev-parse HEAD)
DATE_FMT = +%Y%m%d
ifdef SOURCE_DATE_EPOCH
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" $(DATE_FMT))
else
    BUILD_DATE ?= $(shell date $(DATE_FMT))
endif
# TAG can be provided as an envvar (provided in the .spec file)
TAG           ?= $(shell git describe --tags --exact-match HEAD 2> /dev/null)
# CLOSEST_TAG can be provided as an envvar (provided in the .spec file)
CLOSEST_TAG   ?= $(shell git describe --tags)
# VERSION is inferred from CLOSEST_TAG
# It accepts tags of type `vX.Y.Z`, `vX.Y.Z-(alpha|beta|rc|...)` and produces X.Y.Z
VERSION       := $(shell echo $(CLOSEST_TAG) | sed -E 's/v(([0-9]\.?)+).*/\1/')
TAGS          := development
PROJECT_PATH  := github.com/flavio/kuberlr
KUBERLR_LDFLAGS  = -ldflags "-X=$(PROJECT_PATH)/pkg/kuberlr.Version=$(VERSION) \
														-X=$(PROJECT_PATH)/pkg/kuberlr.BuildDate=$(BUILD_DATE) \
														-X=$(PROJECT_PATH)/pkg/kuberlr.Tag=$(TAG) \
														-X=$(PROJECT_PATH)/pkg/kuberlr.ClosestTag=$(CLOSEST_TAG)"

KUBERLR_DIRS = cmd pkg internal

# go source files
KUBERLR_SRCS = $(shell find $(KUBERLR_DIRS) -type f -name '*.go')

# Define target platforms, image builder and the fully qualified image name.
TARGET_PLATFORMS ?= linux/amd64,linux/arm64,linux/s390x

RUNNER := docker
IMAGE_BUILDER := $(RUNNER) buildx
MACHINE := kwctl

REPO ?= flavio
IMAGE = $(REPO)/kuberlr:$(TAG)
BUILD_ACTION = --load

.PHONY: all
all: install

.PHONY: build
build: go-version-check
	$(GO) build $(KUBERLR_LDFLAGS) -tags $(TAGS) ./cmd/...

.PHONY: install
install: go-version-check
	$(GO) install $(KUBERLR_LDFLAGS) -tags $(TAGS) ./cmd/...

.PHONY: clean
clean:
	$(GO) clean -i ./...
	$(RM) -f ./kuberlr
	$(RM) -rf $(BINPATH)

.PHONY: distclean
distclean: clean
	$(GO) clean -i -cache -testcache -modcache ./...

.PHONY: staging
staging:
	make TAGS=staging install

.PHONY: release
release:
	make TAGS=release install

.PHONY: go-version-check
go-version-check:
	@[ $(GO_VERSION_MAJ) -ge 2 ] || \
		[ $(GO_VERSION_MAJ) -eq 1 -a $(GO_VERSION_MIN) -ge 20 ] || (echo "FATAL: Go version should be >= 1.20.x" ; exit 1 ; )

.PHONY: lint
lint: golangci-lint
	GO111MODULE=on go mod tidy && GO111MODULE=on go mod verify
	# run go fmt
	test -z `$(GOFMT) -l $(KUBERLR_SRCS)` || { $(GOFMT) -d $(KUBERLR_SRCS) && false; }
	# run golangci-lint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

# tests
.PHONY: test
test: test-unit test-bench

.PHONY: test-unit
test-unit:
	$(GO) test -coverprofile=coverage.out $(PROJECT_PATH)/{cmd,pkg,internal}/...

.PHONY: test-unit-coverage
test-unit-coverage: test-unit
	$(GO) tool cover -html=coverage.out

.PHONY: test-bench
test-bench:
	$(GO) test -bench=. $(PROJECT_PATH)/{cmd,pkg,internal}/...

buildx-machine: ## create rancher dockerbuildx machine targeting platform defined by DEFAULT_PLATFORMS.
	@docker buildx ls | grep $(MACHINE) || \
		docker buildx create --name=$(MACHINE) --platform=$(TARGET_PLATFORMS)

image-build: buildx-machine ## build (and load) the container image targeting the current platform.
	$(IMAGE_BUILDER) build -f package/Dockerfile \
		--progress plain --no-cache \
		--builder $(MACHINE) $(IMAGE_ARGS) \
		--build-arg PROJECT_PATH=$(PROJECT_PATH) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg TAG=$(TAG) \
		-t "$(IMAGE)" $(BUILD_ACTION) .
	@echo "Built $(IMAGE)"

generate-mocks:
	mockery

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.7.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
