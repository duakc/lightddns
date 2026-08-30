NAME=lightddns
MODULE=github.com/duakc/lightddns
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)
SCRIPT_WORKDIR=$(shell cd script/goscript && pwd)

BUILD_DIR=build
RELEASE_FILE_DIR=release

BUILD_GOOS ?= *
BUILD_GOARCH ?= *

BUILD_VERSION ?= $(shell git describe --tags --dirty)
BUILD_BRANCH ?= $(shell git branch --show-current)

BUILDINFO_ARGS=--buildinfo_version '$(BUILD_VERSION)' \
	--buildinfo_branch '$(BUILD_BRANCH)'

GOBUILD_TAGS ?=
GOBUILD_ENV ?=
GOBUILD_EXTRA_ARGS ?=
GOBUILD_LDFLAGS ?= \
	-X "$(MODULE)/constant.Version=$(patsubst v%,%,$(BUILD_VERSION))" \
	-X "$(MODULE)/constant.Tags=$(GOBUILD_TAGS)" \
	-X "$(MODULE)/infra/netx/httpx.DefaultUserAgent=$(NAME)/$(BUILD_VERSION) ($(shell go env GOVERSION))"

GOBUILD_ARGS=--gobuild_workdir '$(MAIN_WORKDIR)' \
	--gobuild_output '$(BUILD_DIR)/bin' \
	--gobuild_binary_name '$(NAME)' \
	--gobuild_tags '$(GOBUILD_TAGS)' \
	--gobuild_env '$(GOBUILD_ENV)' \
	--gobuild_ldflags '$(GOBUILD_LDFLAGS)' \
	--gobuild_extra_args '$(GOBUILD_EXTRA_ARGS)'

GO_SCRIPT=GOSCRIPT_DRAFT_SUB_DIR="draft" \
		GOSCRIPT_BUILD_DIR=$(BUILD_DIR) \
		GOSCRIPT_RELEASE_DIR=$(RELEASE_FILE_DIR) \
		go run $(SCRIPT_WORKDIR)/run.go

ifeq ($(origin DOCKER_CLI), undefined)
    DOCKER_CLI := $(shell command -v nerdctl || command -v docker)
endif

.DEFAULT_GOAL=build

.PHONY: all
all: toolchain clean generate test \
	generate-schema build-release-binary build-release-package build-all

.PHONY: build-release
build-release: clean generate generate-schema build-release-binary build-release-package

.PHONY: build-release-package
build-release-package: build-deb build-rpm build-archlinux build-alpine-apk build-openwrt

.PHONY: build-release-binary
build-release-binary: build-all
	find "$(BUILD_DIR)/bin" -type f \
	-not -name '*.gz' \
	-not -name '*.exe' \
	-print0 | xargs -0 -r gzip --best;

.PHONY: test
test: lint
	@go test -v --tags debug ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: generate-schema
generate-schema:
	$(GO_SCRIPT) genschema

.PHONY: toolchain
toolchain:
	@asdf install

.PHONY: fmt
fmt:
	golangci-lint fmt ./...

.PHONY: lint
lint: fmt
	go vet ./...
	golangci-lint run --timeout=30s ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --timeout=30s --fix ./...

.PHONY: clean
clean:
	@rm -rf site/ $(BUILD_DIR) \
		.cache/ .wrangler/
	@go mod tidy

.PHONY: build-all
build-all: override GOBUILD_ARGS += --gobuild_qualified
build-all: generate
	@$(GO_SCRIPT) build $(BUILDINFO_ARGS) $(GOBUILD_ARGS) \
		--goarch '$(BUILD_GOARCH)' \
		--goos '$(BUILD_GOOS)'

.PHONY: build-dev
comma := ,
build-dev: override GOBUILD_TAGS := $(if $(strip $(GOBUILD_TAGS)),$(GOBUILD_TAGS)$(comma),)debug
build-dev: generate
	@$(GO_SCRIPT) build $(BUILDINFO_ARGS) $(GOBUILD_ARGS) \
		--goarch '$(shell go env GOARCH)' \
		--goos '$(shell go env GOOS)'

.PHONY: build
build: generate
	@$(GO_SCRIPT) build $(BUILDINFO_ARGS) $(GOBUILD_ARGS) \
		--goarch '$(shell go env GOARCH)' \
		--goos '$(shell go env GOOS)'

.PHONY: build-docs
build-docs:
	@pip install -r requirements.txt && mkdocs build

.PHONY: build-docker
build-docker:
	$(DOCKER_CLI) build -t $(NAME):latest -f Dockerfile .

.PHONY: build-deb
build-deb: generate
	$(GO_SCRIPT) nfpm $(BUILDINFO_ARGS) $(GOBUILD_ARGS) --format deb --goarch '$(BUILD_GOARCH)'

.PHONY: build-rpm
build-rpm: generate
	$(GO_SCRIPT) nfpm $(BUILDINFO_ARGS) $(GOBUILD_ARGS) --format rpm --goarch '$(BUILD_GOARCH)'

.PHONY: build-archlinux
build-archlinux: generate
	$(GO_SCRIPT) nfpm $(BUILDINFO_ARGS) $(GOBUILD_ARGS) --format archlinux --goarch '$(BUILD_GOARCH)'

.PHONY: build-alpine-apk
build-alpine-apk: generate
	$(GO_SCRIPT) nfpm $(BUILDINFO_ARGS) $(GOBUILD_ARGS) --format alpine.apk --goarch '$(BUILD_GOARCH)'

.PHONY: build-openwrt
build-openwrt: generate
	$(GO_SCRIPT) nfpm $(BUILDINFO_ARGS) $(GOBUILD_ARGS) --format openwrt --goarch '$(BUILD_GOARCH)'

# Nix is declarative: the flake (release/nix, symlinked as ./flake.nix) builds
# the package for the host system. The result lands under build/nix/result.
.PHONY: build-nix
build-nix:
	nix build .#default --print-build-logs -o $(BUILD_DIR)/nix/result
