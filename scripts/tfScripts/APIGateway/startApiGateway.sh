#!/bin/bash
cd /home/group9/CloudVideoConverter/
sudo rm nohup.out
/usr/local/go/bin/go run api-gateway/server/main.go 50051
