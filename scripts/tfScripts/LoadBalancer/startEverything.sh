#!/bin/bash
sudo apt-get install git -y
git clone https://github.com/Frans-Lukas/CloudVideoConverter.git
chmod +x CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh
./CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh
./CloudVideoConverter/scripts/tfScripts/SpawningPool/startAPIGatewayVM.sh