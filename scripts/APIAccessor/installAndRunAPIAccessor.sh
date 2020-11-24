#!/bin/bash
if [ "$#" -ne 5 ]; then
  echo "Usage: $0 API_IP API_PORT {-add|-remove} THISIP THISPORT " >&2
  echo "EX: $0 localhost 1337 -add 132.13.3.7 50041 " >&2
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
