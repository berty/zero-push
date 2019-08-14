BUILD_ENV ?= GO111MODULE=on
PROTOC_OPTS = -I ./vendor/github.com/gogo/protobuf:vendor:.
PROTOC ?= protoc
GOPATH ?= $(HOME)/go
BIN ?= $(GOPATH)/bin/zeropush

$(BIN): generate
	$(BUILD_ENV) go install -v ./cmd/zeropush/...

generate:
	$(PROTOC) $(PROTOC_OPTS) \
		--gofast_out="plugins=grpc:$(GOPATH)/src" \
		./proto/*.proto

.PHONY: lint
lint:
	GO111MODULE=off golangci-lint run --deadline=5m --verbose  ./...
