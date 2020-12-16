#!/bin/bash
sudo apt-get install git -y
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
chmod +x CloudVideoConverter/scripts/terraform/APIGateway/startAPIGatewayVM.sh
./CloudVideoConverter/scripts/terraform/APIGateway/startAPIGatewayVM.sh