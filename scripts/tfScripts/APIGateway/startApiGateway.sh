#!/bin/bash
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"
sudo git pull
go run api-gateway/server/main.go 50051
