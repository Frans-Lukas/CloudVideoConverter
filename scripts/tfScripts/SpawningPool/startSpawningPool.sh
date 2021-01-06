#!/bin/bash
cd /home/group9/CloudVideoConverter

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
echo "starting with IP: '$IP'"
sudo /usr/local/go/bin/go run spawning-pool/client/main.go $IP 50051

