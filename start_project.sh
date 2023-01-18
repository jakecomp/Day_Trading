#!/bin/bash 

./create_docker_network.sh 
./mongodb/start_mongo.sh  

cd backend_api

go build main.go 
go run main.go 

cd ..


