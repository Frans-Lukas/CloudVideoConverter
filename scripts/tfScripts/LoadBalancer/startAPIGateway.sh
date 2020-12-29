#!/bin/bash
cd CloudVideoConverter
sudo git pull
cd ..
chmod +x CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh
./CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh