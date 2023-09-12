GOMOD ?= on
GO ?= GO111MODULE=$(GOMOD) go

#Don't enable mod=vendor when GOMOD is off or else go build/install will fail
GOMODFLAG ?=
ifeq ($(GOMOD), off)
GOMODFLAG=
endif

#retrieve go version details for version check
GO_VERSION     := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_VERSION_MAJ := $(shell echo $(GO_VERSION) | cut -f1 -d'.')
GO_VERSION_MIN := $(shell echo $(GO_VERSION) | cut -f2 -d'.')

GOFMT ?= gofmt
RM = rm

BINPATH       := $(abspath ./bin)
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

# go source files, ignore vendor directory
KUBERLR_SRCS = $(shell find $(KUBERLR_DIRS) -type f -name '*.go')

.PHONY: all
all: install

.PHONY: build
build: go-version-check
	$(GO) build $(GOMODFLAG) $(KUBERLR_LDFLAGS) -tags $(TAGS) ./cmd/...

.PHONY: install
install: go-version-check
	$(GO) install $(GOMODFLAG) $(KUBERLR_LDFLAGS) -tags $(TAGS) ./cmd/...

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
lint: deps
	# explicitly enable GO111MODULE otherwise go mod will fail
	GO111MODULE=on go mod tidy && GO111MODULE=on go mod vendor && GO111MODULE=on go mod verify
	# run go vet
	$(GO) vet ./...
	# run go fmt
	test -z `$(GOFMT) -l $(KUBERLR_SRCS)` || { $(GOFMT) -d $(KUBERLR_SRCS) && false; }
	# run golint
	golint -set_exit_status cmd/... internal/... pkg/...

.PHONY: deps
deps:
	go get -u golang.org/x/lint/golint

# tests
.PHONY: test
test: test-unit test-bench

.PHONY: test-unit
test-unit:
	$(GO) test $(GOMODFLAG) -coverprofile=coverage.out $(PROJECT_PATH)/{cmd,pkg,internal}/...

.PHONY: test-unit-coverage
test-unit-coverage: test-unit
	$(GO) tool cover -html=coverage.out

.PHONY: test-bench
test-bench:
	$(GO) test $(GOMODFLAG) -bench=. $(PROJECT_PATH)/{cmd,pkg,internal}/...
