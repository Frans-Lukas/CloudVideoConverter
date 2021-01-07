#!/bin/bash
cd ../LoadBalancerWithoutAPIGateWay
../LoadBalancer/terraform destroy --auto-approve -var-file="../variables.tfvars" -var 'instance_count=3'
cd ../LoadBalancer
./terraform destroy --auto-approve -var-file="../variables.tfvars"
./terraform init
./terraform apply --auto-approve -var-file="../variables.tfvars"
