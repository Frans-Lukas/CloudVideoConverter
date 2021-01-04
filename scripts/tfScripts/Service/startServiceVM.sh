#!/bin/bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi
cd /home/group9/CloudVideoConverter/scripts/tfScripts/Service
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='$1'' -var-file="../variables.tfvars"