FROM znly/protoc:latest as proto
WORKDIR /root/go/src/berty.tech/zero-push

RUN apk add --no-cache make
ENV PROTOC_OPTS="-I/protobuf -I."
COPY . .
RUN make clean generate

FROM golang:1.12-alpine3.10 as builder
COPY --from=proto /root/go/src/berty.tech/zero-push /go/src/berty.tech/zero-push
WORKDIR /go/src/berty.tech/zero-push

RUN apk add --no-cache make git
RUN go mod download && make build

FROM alpine:3.10
COPY --from=builder /go/bin/zeropush /bin/zeropush

ENTRYPOINT ["/bin/zeropush"]
