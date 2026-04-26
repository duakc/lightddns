.PHONY: all
all: clean generate test build-all

.PHONY: test
test: lint
	@go test -v ./...

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
	rm -rf ./bin/build/

.PHONY: build-all
build-all:
	@go run -C script/goscript . build -verbose -all

.PHONY: build
build:
	@go run -C script/goscript . build -verbose