#!/bin/bash

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
while true; do
    instances="$(gcloud compute instances list --format='table(name)')"
    while IFS= read -r line
    do
        if [[ "$line" == "START" ]]; then
            echo $line
        elif [["$line" == *"load-balancer"*]]; then
            echo "killing '$line'"
        elif [["$line" == *"service-provider"*]]; then
            echo "killing '$line'"
        elif [["$line" == *"api-gateway"*]]; then
            echo "ignoring '$line'"
        elif [["$line" == *"spawning-pool"*]]; then
            echo "ignoring '$line'"
        fi
    done < <(printf '%s\n' "$instances")
    sleep 15
done
#gcloud compute instances delete
