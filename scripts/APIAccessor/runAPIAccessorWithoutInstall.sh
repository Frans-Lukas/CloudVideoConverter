#!/bin/bash
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 API_IP API_PORT THISPORT " >&2
  echo "EX: $0 localhost 1337 50041 " >&2
  exit 1
fi
cd CloudVideoConverter
export PATH=$PATH:/usr/local/go/bin
git pull
publicIp=$(curl -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip)
go run api-gateway/client/main.go $1 $2 $publicIp $3
