#!/bin/bash

gcloud auth activate-service-account fast-blueprint-296210@appspot.gserviceaccount.com --key-file=/tmp/SSDNIA.json
gcloud compute instances list
#gcloud compute instances delete
