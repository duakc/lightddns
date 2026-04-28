NAME="lightddns"
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)
SCRIPT_WORKDIR=$(shell cd script/goscrip && pwd)

GIT_VERSION=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git branch --show-current)
BUILD_OUTPUT="./bin/build"

GO_SCRIPT=go run -C $(SCRIPT_WORKDIR) .

GO_BUILD=$(GO_SCRIPT) build -verbose \
			-version $(GIT_VERSION) -branch $(GIT_BRANCH)\
			-workdir $(MAIN_WORKDIR) -binary $(NAME)\
			-output $(BUILD_OUTPUT)


.PHONY: all
all: clean generate test build

.PHONY: test
test: lint
	@go test -v -tags debug ./...

.PHONY: generate
generate:
	@go generate ./...

.PHONY: generate-doc
generate-doc:
	@$(GO_SCRIPT) schema

.PHONY: toolchain
toolchain:
	@asdf install

.PHONY: fmt
fmt:
	@golangci-lint fmt

.PHONY: lint
lint: fmt
	@golangci-lint run

.PHONY: clean
clean:
	rm -rf $(BUILD_OUTPUT)
	go mod tidy
	cd script/goscript && go mod tidy

.PHONY: build-all
build-all:
	@$(GO_BUILD) -all

.PHONY: build-dev
build-dev:
	@$(GO_BUILD) -tags debug

.PHONY: build
build:
	@$(GO_BUILD)

