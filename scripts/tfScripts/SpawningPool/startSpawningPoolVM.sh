#!/bin/bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi
cd scripts/tfScripts/SpawningPool
export PATH=$PATH:/Home/staff/pirat/IdeaProjects/CloudVideoConverter/scripts/tfScripts/LoadBalancer
terraform init
terraform apply -input=false -auto-approve -var 'instance_count='$1''