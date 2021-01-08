#!/bin/bash
cd CloudVideoConverter || echo "CloudVideoConverter does not exist"


#download video to localStorage
IP=$(curl https://ipinfo.io/ip)
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
API_IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
NAME=$(curl http://metadata.google.internal/computeMetadata/v1/instance/hostname -H Metadata-Flavor:Google)

while true; do
    /usr/local/go/bin/go run converter/server/main.go ${IP} 50053 "$API_IP:50051" $NAME
    echo "restarting service"
    sleep 15
done
