#!/bin/bash

#protoc --proto_path=api/proto --proto_path=third_party --go_out=plugins=grpc:api-gateway/generated api-gateway.proto
protoc --proto_path=api/proto --proto_path=third_party --go_out=api-gateway/generated --go_opt=paths=source_relative --go-grpc_out=api-gateway/generated --go-grpc_opt=paths=source_relative api-gateway.proto