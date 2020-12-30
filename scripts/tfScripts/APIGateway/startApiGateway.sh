#!/bin/bash
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"

/usr/local/go/bin/go run api-gateway/server/main.go 50051
