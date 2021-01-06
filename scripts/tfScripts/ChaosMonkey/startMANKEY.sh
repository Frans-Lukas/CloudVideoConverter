#!/bin/bash
../LoadBalancer/terraform destroy --auto-approve -var-file="../variables.tfvars"
../LoadBalancer/terraform init
../LoadBalancer/terraform apply --auto-approve -var-file="../variables.tfvars"