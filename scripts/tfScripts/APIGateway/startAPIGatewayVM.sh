#!/bin/bash
cd /home/group9/CloudVideoConverter/scripts/tfScripts/APIGateway
sudo terraform destroy -input=false -auto-approve -var 'instance_count='1'' -var-file="/home/group9/CloudVideoConverter/scripts/tfScripts/variables.tfvars"
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='1'' -var-file="/home/group9/CloudVideoConverter/scripts/tfScripts/variables.tfvars"
