#!/bin/bash
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"
go run api-gateway/server/main.go 50051
