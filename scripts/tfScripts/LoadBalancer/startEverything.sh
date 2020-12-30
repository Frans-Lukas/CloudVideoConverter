#!/bin/bash
./terraform destroy --auto-approve -var-file="../variables.tfvars"
./terraform init
./terraform apply --auto-approve -var-file="../variables.tfvars"
#read -p "Enter ip of load-balancer-0 VM: " ip
#echo $ip
#ssh-add /tmp/id_rsa
#ssh franslukas@$ip
