#!/bin/bash
./terraform destroy --auto-approve -var-file="../variables.tfvars"
./terraform init
./terraform apply --auto-approve -var-file="../variables.tfvars"