#!/bin/bash

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
while true; do
    instances="$(gcloud compute instances list --format='table(name)')"
    while IFS= read -r line
    do
        if [[ $line == *"load-balancer"* ]]; then
            echo "killing '$line'"
            if (( RANDOM % 100 <= 25 )); then
                gcloud compute instances delete $line --zone=europe-north1-a -q
                break
            fi
        elif [[ $line == *"service-provider"* ]]; then
            echo "killing '$line'"
            if (( RANDOM % 100 <= 25 )); then
                gcloud compute instances delete $line --zone=europe-north1-a -q
                break
            fi
        elif [[ $line == *"api-gateway"* ]]; then
            echo "ignoring '$line'"
        elif [[ $line == *"spawning-pool"* ]]; then
            echo "ignoring '$line'"
        fi
    done < <(printf '%s\n' "$instances")
    sleep RANDOM % 25 + 5
done
#gcloud compute instances delete
