DOCKER_TAG ?= v1.0.0

BUILD_ENV ?= GO111MODULE=on
PROTOC_OPTS ?= -I ./vendor/github.com/gogo/protobuf:vendor:.
PROTOC ?= protoc
PROTO_FILES = ./proto/push/push.proto ./proto/service/service.proto
PROTO_GEN = $(PROTO_FILES:.proto=.pb.go)
GOPATH ?= $(HOME)/go

DOCKER_NAME = bertychat/zeropush

.PHONY: all
all: generate build

.PHONY: build
build:
	$(BUILD_ENV) go install -v ./cmd/...

.PHONY: generate
generate: $(PROTO_GEN)

%.pb.go: %.proto
	$(PROTOC) $(PROTOC_OPTS) --gofast_out="plugins=grpc:$(GOPATH)/src" $<

.PHONY: clean
clean:
	rm $(PROTO_GEN)

.PHONY: lint
lint:
	GO111MODULE=off golangci-lint run --deadline=5m --verbose  ./...

.PHONY: docker.build
docker.build:
	docker build -t $(DOCKER_NAME):$(DOCKER_TAG) .

.PHONY: docker.publish
docker.publish: docker.build
	docker push $(DOCKER_NAME):$(DOCKER_TAG)
