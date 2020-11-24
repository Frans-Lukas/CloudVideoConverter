#!/bin/bash
if [ "$#" -ne 4 ]; then
  echo "Usage: $0 API_IP API_PORT THISIP THISPORT" >&2
  exit 1
fi
cd CloudVideoConverter
export PATH=$PATH:/usr/local/go/bin
git pull
go run api-gateway/client/main.go $1 $2 $3 $4
