#!/bin/bash
cd /home/group9/CloudVideoConverter/
sudo git checkout .
sudo git pull
sudo chmod -R +x .
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
/usr/local/go/bin/go run load-balancer/server/main.go 50052 "$IP:50051"

