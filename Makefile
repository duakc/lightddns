NAME="lightddns"
MAIN_WORKDIR=$(shell cd cmd/$(NAME) && pwd)

GIT_VERSION=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git branch --show-current)
BUILD_OUTPUT="./bin/build"

GO_BUILD=go run -C script/goscript . build -verbose \
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

.PHONY: build-all
build-all:
	@$(GO_BUILD) -all

.PHONY: build-dev
build-dev:
	@$(GO_BUILD) -tags debug

.PHONY: build
build:
	@$(GO_BUILD)

