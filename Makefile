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


.PHONY: all
all: clean generate test build

.PHONY: test
test: lint
	@go test -v -tags debug ./...

.PHONY: generate
generate:
	@go generate ./...
	@$(GO_SCRIPT) genschema

.PHONY: toolchain
toolchain:
	@asdf install

.PHONY: fmt
fmt:
	@golangci-lint fmt

.PHONY: lint
lint: fmt
	@golangci-lint run

.PHONY: lint-fix
lint-fix:
	@golangci-lint run --fix

.PHONY: clean
clean:
	rm -rf $(BUILD_OUTPUT) site/ .cache/
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
	$(GO_BUILD)

.PHONY: build-docs
build-docs:
	@pip install -r requirements.txt
	@mkdocs build

