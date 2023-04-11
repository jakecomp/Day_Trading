###You will probs need latest docker and minikube versions so update.
#until I move it all into the docker buildblkz container

##all done from top level folder

docker compose build

minikube start --nodes=1 --driver=docker --embed-certs  --memory=max --cpus=max --addons=ingress,ingress-dns,volumesnapshots,csi-hostpath-driver --service-cluster-ip-range="10.9.0.0/24"

##if this doesnt work for you, your system might name the images diff than mine, so just change them
minikube image load seng_468-backend seng_468-frontend seng_468-mongo_db seng_468-log_service seng_468-quote_server seng_468-quote_queuer seng_468-trigger_service seng_468-worker

kubectl apply -f ./yamlnew/.

##wait for a lil bit then portforward so cli_app can access
kubectl port-forward service/backend 8000:8000 -n seng-trade-app 

##test
./cli_app/app < ./cli_app/workflows/user10.txt 


###need to cleanup files still very messy lol

##play around change auto scale for w/e deployments

kubectl autoscale deployment/backend-deploy --min=1 --max=3
minikube start --nodes=2 //or more? extra nodes