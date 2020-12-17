#!/bin/bash
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
cd CloudVideoConverter/scripts/tfScripts/SpawningPool
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='1''