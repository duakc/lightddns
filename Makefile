NAME="lightddns"
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)
SCRIPT_WORKDIR=$(shell cd script/goscript && pwd)

BUILD_DIR="build"
RELEASE_FILE_DIR="release"

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
	generate-schema build-all build-deb build-rpm

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
build-all: build
	@$(GO_SCRIPT) build --all

.PHONY: build-dev
build-dev: generate
	@$(GO_SCRIPT) build --tags debug

.PHONY: build
build: generate
	@$(GO_SCRIPT) build

.PHONY: build-docs
build-docs:
	@pip install -r requirements.txt && mkdocs build

.PHONY: build-docker
build-docker:
	$(DOCKER_CLI) build -t $(NAME):latest -f Dockerfile .

.PHONY: build-deb
build-deb:
	$(GO_SCRIPT) nfpm --format deb --goarch '*'

.PHONY: build-rpm
build-rpm:
	$(GO_SCRIPT) nfpm --format rpm --goarch '*'

.PHONY: build-archlinux
build-nfpm-archlinux:
	$(GO_SCRIPT) nfpm --format archlinux

.PHONY: build-openwrt
build-openwrt:
	$(GO_SCRIPT) nfpm --format openwrt

.PHONY: build-openwrt-all
build-openwrt-all:
	$(GO_SCRIPT) nfpm --format openwrt --all

# Nix is declarative: the flake (release/nix, symlinked as ./flake.nix) builds
# the package for the host system. The result lands under build/nix/result.
.PHONY: build-nix
build-nix:
	nix build .#default --print-build-logs -o $(BUILD_DIR)/nix/result