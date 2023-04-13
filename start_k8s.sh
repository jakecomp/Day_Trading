#!/bin/bash 
#docker compose build 
minikube start --nodes=1 --driver=docker --embed-certs --memory=5000mb --cpus=4 --addons=ingress,ingress-dns,volumesnapshots,csi-hostpath-driver --service-cluster-ip-range="10.9.0.0/24" 
minikube image load seng_468-backend seng_468-frontend seng_468-mongo_db seng_468-log_service seng_468-quote_server seng_468-quote_queuer seng_468-trigger_service seng_468-worker 
echo "Please Wait Applying Configuration"
minikube kubectl -- apply -f ./yamlnew/.
echo "Finished Applying Configuration"
