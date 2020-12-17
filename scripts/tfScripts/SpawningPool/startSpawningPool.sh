#!/bin/bash
sudo apt-get install wget -y
sudo apt-get install git -y
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go get -u google.golang.org/grpc
cd CloudVideoConverter
mkdir localStorage
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
IP=$(gcloud compute instances describe load-balancer-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
go run load-balancer/client/main.go 50052 $IP 50052

