# SENG 468 Day Trading App Spring 2023
This repo holds our group project for the SENG 468 Day Trading App 

## HOW TO START

The easiest way to run our application is via docker-compose. In the 
root directory of our project run the following

```bash
docker-compose up --build -d
``` 

Verfiy all services up and healthy by running 

```bash
docker ps -a
```  

## HOW TO RUN WORKLOAD 

Once all the services are up and running navigate to the cli_app directory. In that directory follow 
the instructions on the README file. 


## Run System on Kubernetes
You will probs need latest docker and minikube versions so update.
until I move it all into the docker buildblkz container

All commands done from top level folder

Build Images

```bash
docker compose build
```  
run start_k8s.sh to initalize/start cluster

wait for a little while so services can start, and then portforward backend service so cli_app can access it

```bash
kubectl port-forward service/backend 8000:8000 -n seng-trade-app 
```  

test workflow

```bash
./cli_app/app < ./cli_app/workflows/user10.txt 
```  

play around; change auto scale for w/e deployments

```bash
kubectl autoscale deployment/backend-deploy --min=1 --max=3
minikube start --nodes=2 //or more? extra nodes
``` 

To test fault tolerance run chaos_test.sh

This will take down a pod every 30s by default