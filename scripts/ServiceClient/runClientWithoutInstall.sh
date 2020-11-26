#!/bin/bash
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 IP PORT" >&2
  exit 1
fi
cd CloudVideoConverter
export PATH=$PATH:/usr/local/go/bin
git pull
go run load-balancer/client/main.go $1 $2