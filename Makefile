GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
APP?=helloworld
VERSION?=$(shell git describe --tags --always 2>/dev/null || echo dev)

.PHONY: init
# init env
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

.PHONY: config
# generate internal proto
config:
	buf generate --template buf.gen.config.yaml

.PHONY: api
# generate api proto
api:
	buf generate --template buf.gen.yaml

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Name=$(APP) -X main.Version=$(VERSION)" -o ./bin/$(APP) ./cmd/$(APP)

.PHONY: run
# run locally
run:
	go run ./cmd/$(APP) -conf ./configs

.PHONY: test
# run tests
test:
	go test ./...

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	$(MAKE) api
	$(MAKE) config
	$(MAKE) generate

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
