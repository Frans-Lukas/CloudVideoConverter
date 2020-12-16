#!/bin/bash
#mkdir ~/.ssh/cloud
#ssh-keygen -b 2048 -f /tmp/id_rsa -t rsa -q -N ""
sudo apt-get update 
sudo apt-get install unzip -y
wget https://releases.hashicorp.com/terraform/0.13.5/terraform_0.13.5_linux_amd64.zip
unzip terraform_0.13.5_linux_amd64.zip 
sudo mv terraform /usr/local/bin
sudo apt-get install git -y
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
cd CloudVideoConverter/scripts/tfScripts/APIGateway
sudo terraform init
sudo terraform apply -input=false -auto-approve -var 'instance_count='1''