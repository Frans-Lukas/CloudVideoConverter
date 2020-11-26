#!/bin/bash
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 PORT" >&2
  exit 1
fi
cd CloudVideoConverter
export PATH=$PATH:/usr/local/go/bin
git pull
go run load-balancer/server/main.go $1
