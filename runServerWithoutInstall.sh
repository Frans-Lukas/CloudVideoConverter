#!/bin/bash
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 PORT" >&2
  exit 1
fi
cd CloudVideoConverter
git pull
go run server/main.go $1
