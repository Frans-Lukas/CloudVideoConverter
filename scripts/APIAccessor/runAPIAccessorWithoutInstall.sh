#!/bin/bash
if [ "$#" -ne 5 ]; then
  echo "Usage: $0 API_IP API_PORT {-add|-remove} THISIP THISPORT " >&2
  echo "EX: $0 localhost 1337 -add 132.13.3.7 50041 " >&2
  exit 1
fi
cd CloudVideoConverter
export PATH=$PATH:/usr/local/go/bin
git pull
go run api-gateway/client/main.go $1 $2 $4 $5 $3
