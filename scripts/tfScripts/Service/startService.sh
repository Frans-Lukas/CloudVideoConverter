#!/bin/bash
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"


#download video to localStorage
IP=$(curl https://ipinfo.io/ip)
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
API_IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
/usr/local/go/bin/go run converter/server/main.go ${IP} 50053 "$API_IP:50051"
