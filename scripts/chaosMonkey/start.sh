#!/bin/bash

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
instances=$(gcloud compute instances list --format='table(name)')
#gcloud compute instances delete
