GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
INTERNAL_PROTO_FILES=$(shell find ginproto -name *.proto)
API_PROTO_FILES=$(shell find example/api -name *.proto)
API_PB_FILES=$(shell find example/api -name *pb.go)

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./example/api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./example/api \
 	       --go-gin_out=paths=source_relative:./example/api \
 	       --openapi_out==paths=source_relative:. --openapi_opt=enum_type=string\
 	       --validate_out=paths=source_relative,lang=go:./example/api \
		   $(API_PROTO_FILES) \
