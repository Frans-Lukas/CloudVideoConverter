#!/bin/bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi
terraform init
terraform apply --auto-approve -var-file="../variables.tfvars" -var 'instance_count='$1''