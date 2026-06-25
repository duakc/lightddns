NAME="lightddns"
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)
SCRIPT_WORKDIR=$(shell cd script/goscript && pwd)

BUILD_DIR="./build"
RELEASE_FILE_DIR="./release"

GO_SCRIPT=go run $(SCRIPT_WORKDIR)/run.go

GO_BUILD=$(GO_SCRIPT) build --verbose \
			--workdir $(MAIN_WORKDIR) --binary $(NAME)

ifeq ($(origin DOCKER_CLI), undefined)
    DOCKER_CLI := $(shell command -v nerdctl || command -v docker)
endif

.DEFAULT_GOAL=build

.PHONY: all
all: toolchain clean generate test \
	generate-schema build-all build-deb-all build-rpm-all

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
	@$(GO_BUILD) --all

.PHONY: build-dev
build-dev: generate
	@$(GO_BUILD) --tags debug

.PHONY: build
build: generate
	@$(GO_BUILD)

.PHONY: build-docs
build-docs:
	@pip install -r requirements.txt && mkdocs build

.PHONY: build-docker
build-docker:
	$(DOCKER_CLI) build -t $(NAME):latest -f Dockerfile .

.PHONY: build-deb
build-deb:
	$(GO_SCRIPT) deb --verbose

.PHONY: build-deb-all
build-deb-all:
	$(GO_SCRIPT) deb --all --verbose

.PHONY: build-rpm
build-rpm:
	$(GO_SCRIPT) rpm --verbose

.PHONY: build-rpm-all
build-rpm-all:
	$(GO_SCRIPT) rpm --all --verbose