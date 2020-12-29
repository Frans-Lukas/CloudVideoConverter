#!/bin/bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi
cd CloudVideoConverter/scripts/terraform/LoadBalancer
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='$1''