NAME="lightddns"
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)
SCRIPT_WORKDIR=$(shell cd script/goscript && pwd)

GIT_VERSION=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git branch --show-current)
BUILD_OUTPUT="./build"

GO_SCRIPT=go run $(SCRIPT_WORKDIR)/run.go

GO_BUILD=$(GO_SCRIPT) build -verbose \
			-version $(GIT_VERSION) -branch $(GIT_BRANCH)\
			-workdir $(MAIN_WORKDIR) -binary $(NAME)\
			-output $(BUILD_OUTPUT)

ifeq ($(origin DOCKER_CLI), undefined)
    DOCKER_CLI := $(shell command -v nerdctl || command -v docker)
endif

.DEFAULT_GOAL=build

.PHONY: all
all: toolchain clean generate test \
	generate-schema build-all

.PHONY: test
test: lint
	@go test -v -tags debug ./...

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
	@golangci-lint fmt ./...

.PHONY: lint
lint: fmt
	@go vet ./...
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	@golangci-lint run --fix ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_OUTPUT) site/ .cache/
	go mod tidy

.PHONY: build-all
build-all: build
	@$(GO_BUILD) -all

.PHONY: build-dev
build-dev: generate
	@$(GO_BUILD) -tags debug

.PHONY: build
build: generate
	$(GO_BUILD)

.PHONY: build-docs
build-docs:
	@pip install -r requirements.txt && mkdocs build

.PHONY: build-docker
build-docker: _check-docker
	$(DOCKER_CLI) build -t $(NAME):latest -f Dockerfile .

.PHONY: _check-docker
_check-docker:
	@if [ -z "$(DOCKER_CLI)" ]; then \
		echo "Please set DOCKER_CLI manually or install nerdctl/docker."; \
		exit 1; \
	fi