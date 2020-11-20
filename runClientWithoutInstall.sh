#!/bin/bash
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 IP PORT" >&2
  exit 1
fi
cd CloudVideoConverter
git pull
go run client/main.go $1 $2