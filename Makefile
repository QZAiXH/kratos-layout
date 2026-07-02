GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION?=$(shell git describe --tags --always 2>/dev/null || echo dev)
CMD_DIR?=$(shell find ./cmd -mindepth 1 -maxdepth 1 -type d | head -1)
PROTO_IDE_DEPS_DIR?=.proto-deps
PROTO_IDE_GOOGLEAPIS_REF?=$(shell awk '/name: buf.build\/googleapis\/googleapis/{found=1} found && /commit:/ {print $$2; exit}' buf.lock)

.PHONY: init
# init env
init:
	go install github.com/google/wire/cmd/wire@v0.7.0
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install entgo.io/ent/cmd/ent@v0.14.6
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0

.PHONY: config
# generate internal proto
config:
	go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.config.yaml

.PHONY: api
# generate api proto
api:
	go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf.gen.yaml

.PHONY: openapi
# generate OpenAPI 3.1 docs
openapi:
	go run ./scripts/openapi

.PHONY: ent
# generate ent code
ent:
	go generate ./internal/data/ent

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ $(CMD_DIR)

.PHONY: run
# run locally
run:
	go run $(CMD_DIR) -conf ./configs

.PHONY: test
# run tests
test:
	go test ./...

.PHONY: lint
# run lint
lint:
	golangci-lint run

.PHONY: proto-ide
# export proto deps for IDE import paths
proto-ide:
	@if [ -z "$(PROTO_IDE_GOOGLEAPIS_REF)" ]; then echo "buf.lock missing googleapis dependency; run buf dep update first"; exit 1; fi
	rm -rf $(PROTO_IDE_DEPS_DIR)
	buf export buf.build/googleapis/googleapis:$(PROTO_IDE_GOOGLEAPIS_REF) --output=$(PROTO_IDE_DEPS_DIR)

.PHONY: generate
# generate
generate:
	go run github.com/google/wire/cmd/wire@v0.7.0 ./cmd/...
	go mod tidy

.PHONY: all
# generate all
all:
	$(MAKE) api
	$(MAKE) openapi
	$(MAKE) config
	$(MAKE) ent
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
