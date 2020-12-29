#!/bin/bash
cd scripts/tfScripts/SpawningPool
export PATH=$PATH:/Home/staff/pirat/IdeaProjects/CloudVideoConverter/scripts/tfScripts/LoadBalancer
terraform init
terraform destroy -input=false -auto-approve