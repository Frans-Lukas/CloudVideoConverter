#!/bin/bash

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
instances=$(gcloud compute instances list --format='table(name)')
while read -r line
do
    echo "$line"
done < <(instances)
#gcloud compute instances delete
