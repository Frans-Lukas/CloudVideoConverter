#!/bin/bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi
cd CloudVideoConverter
git pull
cd CloudVideoConverter/scripts/tfScripts/Service
sudo terraform init
sudo terraform destroy -input=false -auto-approve -target google_compute_instance.vm_instance[$1] 