#!/usr/bin/env bash
if [ $# -ne 1 ]; then
    echo "The number of arguments passed is incorrect"
    exit 1
fi

cd /home/group9/CloudVideoConverter/scripts/tfScripts/LoadBalancerWithoutAPIGateWay
./startLoadBalancerVMs.sh $1