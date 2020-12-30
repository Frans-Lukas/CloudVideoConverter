#!/bin/bash
cd CloudVideoConverter
sudo git pull
cd scripts/tfScripts/APIGateway
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='1'' -var-file="../variables.tfvars"