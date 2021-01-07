#!/bin/bash
cd /home/group9/CloudVideoConverter/
sudo git checkout .
sudo git pull
sudo chmod -R +x .
sudo rm localStorage/*
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
MY_IP=$(curl https://ipinfo.io/ip)

echo $MY_IP

while true
do
    /usr/local/go/bin/go run load-balancer/server/main.go $MY_IP 50052 50054 "$IP:50051"
    echo "Server 'load balancer' crashed with exit code $?.  Respawning.."
    sleep 5
done

