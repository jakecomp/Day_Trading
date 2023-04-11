#!/usr/bin/env bash
#   Copyright IBM Corporation 2020
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# Invoke as pushimages.sh <registry_url> <registry_namespace> <container_runtime>

REGISTRY_URL=docker.io
REGISTRY_NAMESPACE=dylanjkemp
CONTAINER_RUNTIME=docker
if [ "$#" -gt 1 ]; then
  REGISTRY_URL=$1
  REGISTRY_NAMESPACE=$2
fi
if [ "$#" -eq 3 ]; then
    CONTAINER_RUNTIME=$3
fi
if [ "${CONTAINER_RUNTIME}" != "docker" ] && [ "${CONTAINER_RUNTIME}" != "podman" ]; then
   echo 'Unsupported container runtime passed as an argument for pushing the images: '"${CONTAINER_RUNTIME}"
   exit 1
fi
# Uncomment the below line if you want to enable login before pushing
#${CONTAINER_RUNTIME} login ${REGISTRY_URL}
#docker login -u="dylanjkemp+stockappseng468" -p="0EKAGH5LR8Y4AGKRU8K7XIV0UKD309ML4V5N8S2W3WE59MWC8ZSDWMQJEQU5V6QP" quay.io

echo 'pushing image worker-service'
${CONTAINER_RUNTIME} tag worker-service ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_worker-service
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_worker-service

echo 'pushing image backend:latest'
${CONTAINER_RUNTIME} tag backend:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_backend:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_backend:latest

echo 'pushing image quote-server'
${CONTAINER_RUNTIME} tag quote-server ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_quote-server
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_quote-server

echo 'pushing image queue-service'
${CONTAINER_RUNTIME} tag queue-service ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_queue-service
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_queue-service

echo 'pushing image m'
${CONTAINER_RUNTIME} tag m ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_cli
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_cli

echo 'pushing image log-service'
${CONTAINER_RUNTIME} tag log-service ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_log-service
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_log-service

echo 'pushing image backend-api'
${CONTAINER_RUNTIME} tag backend-api ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_backend-api
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_backend-api

echo 'pushing image worker:latest'
${CONTAINER_RUNTIME} tag worker:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_worker:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_worker:latest

echo 'pushing image quote-server:latest'
${CONTAINER_RUNTIME} tag quote-server:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_quote-server:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_quote-server:latest

echo 'pushing image quote-cache:latest'
${CONTAINER_RUNTIME} tag quote-cache:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_quote-cache:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_quote-cache:latest

echo 'pushing image mongo-db:latest'
${CONTAINER_RUNTIME} tag mongo-db:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_mongo-db:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_mongo-db:latest

echo 'pushing image log-service:latest'
${CONTAINER_RUNTIME} tag log-service:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_log-service:latest
${CONTAINER_RUNTIME} push /${REGISTRY_NAMESPACE}/stockapp_log-service:latest

echo 'pushing image frontend:latest'
${CONTAINER_RUNTIME} tag frontend:latest ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/stockapp_frontend:latest
${CONTAINER_RUNTIME} push ${REGISTRY_NAMESPACE}/stockapp_frontend:latest

echo 'done'
