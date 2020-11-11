#!/bin/bash

protoc --proto_path=api/proto --proto_path=third_party --go_out=plugins=grpc:generated videoconverter.proto
