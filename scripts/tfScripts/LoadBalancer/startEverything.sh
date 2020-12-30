#!/bin/bash
./terraform destroy --auto-approve -var-file="../variables.tfvars"
./terraform init
./terraform apply --auto-approve -var-file="../variables.tfvars"
read -p "Enter ip of load-balancer-0 VM: " ip
echo $ip
ssh-add /tmp/id_rsa
ssh -oStrictHostKeyChecking=no -t -t group9@$ip << EOF
  cd /home/group9/CloudVideoConverter
  sudo git checkout .
  sudo git pull
  sudo chmod -R +x /home/group9/*
  cd /home/group9/CloudVideoConverter/scripts/tfScripts/APIGateway
  sudo /home/group9/CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh
  /home/group9/CloudVap  ideoConverter/scripts/tfScripts/LoadBalancer/startLoadBalancer.sh &
EOF
