#!/bin/bash

sudo apt-get install wget -y
sudo apt-get install git -y
wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
sudo apt-get install ffmpeg -y
export PATH=$PATH:/usr/local/go/bin
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"
mkdir localStorage
#download video to localStorage
IP=$(curl https://ipinfo.io/ip)
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
API_IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
export GOOGLE_APPLICATION_CREDENTIALS=/tmp/SSDNIA.json
go run converter/server/main.go ${IP} 50053 "$API_IP:50051"
