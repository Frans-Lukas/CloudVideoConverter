#!/bin/bash
export GOOGLE_APPLICATION_CREDENTIALS=/tmp/SSDNIA.json
cd CloudVideoConverter
sudo git pull
cd ..
sudo chmod +x CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh
./CloudVideoConverter/scripts/tfScripts/APIGateway/startAPIGatewayVM.sh