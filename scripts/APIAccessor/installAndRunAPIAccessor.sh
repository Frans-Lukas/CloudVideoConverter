#!/bin/bash
if [ "$#" -ne 4 ]; then
  echo "Usage: $0 API_IP API_PORT THISIP THISPORT" >&2
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
go run api-gateway/client/main.go $1 $2 $3 $4
