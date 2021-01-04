#!/bin/bash
cd /home/group9/CloudVideoConverter/scripts/tfScripts/APIGateway
sudo terraform destroy -input=false -auto-approve -var 'instance_count='1'' -var-file="/home/group9/CloudVideoConverter/scripts/tfScripts/variables.tfvars"
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='1'' -var-file="/home/group9/CloudVideoConverter/scripts/tfScripts/variables.tfvars"
gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
IP=$(gcloud compute instances describe api-gateway-0 --format='get(networkInterfaces[0].accessConfigs[0].natIP)' --zone=europe-north1-a)
echo "IP='$IP'"
sudo pkill ssh-agent
eval "$(ssh-agent)"
ssh-add /tmp/id_rsa
ssh -oStrictHostKeyChecking=no -t -t group9@$IP << EOF
  cd /home/group9/CloudVideoConverter
  sudo git checkout .
  sudo git pull
  sudo chmod -R +x /home/group9/*
  cd /home/group9/CloudVideoConverter/
  sudo nohup /home/group9/CloudVideoConverter/scripts/tfScripts/APIGateway/startApiGateway.sh &
  sleep 1
  exit
EOF