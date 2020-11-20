#!/bin/bash

sudo apt-get install wget -y
sudo apt-get install git -y
sudo apt-get install ffmpeg -y
wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go get -u google.golang.org/grpc
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
cd CloudVideoConverter
mkdir localStorage
go run server/main.go 50051
