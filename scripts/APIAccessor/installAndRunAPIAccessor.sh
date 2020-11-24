#!/bin/bash
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 API_IP API_PORT THISPORT " >&2
  echo "EX: $0 localhost 1337 50041 " >&2
  exit 1
fi

sudo apt-get install wget -y
sudo apt-get install git -y
wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go get -u google.golang.org/grpc
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
cd CloudVideoConverter
#download video to localStorage

publicIp=$(curl -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip)
go run api-gateway/client/main.go $1 $2 $publicIp $3
