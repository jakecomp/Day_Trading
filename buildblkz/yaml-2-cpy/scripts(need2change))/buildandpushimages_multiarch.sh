#!/usr/bin/env bash
#   Copyright IBM Corporation 2021
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

if [[ "$(basename "$PWD")" != 'scripts' ]] ; then
  echo 'please run this script from the "scripts" directory'
  exit 1
fi

cd .. # go to the parent directory so that all the relative paths will be correct

REGISTRY_URL=quay.io
REGISTRY_NAMESPACE=stockapp
PLATFORMS="linux/amd64,linux/arm64,linux/s390x,linux/ppc64le"
if [ "$#" -gt 1 ]; then
  REGISTRY_URL=$1
  REGISTRY_NAMESPACE=$2
fi
if [ "$#" -eq 3 ]; then
  PLATFORMS=$3
fi
# Uncomment the below line if you want to enable login before pushing
# docker login ${REGISTRY_URL}

echo 'building and pushing image backend:latest'
cd source/seng-468---copy/seng_468 - Copy/backend_api
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/backend:latest .
cd -

echo 'building and pushing image frontend:latest'
cd source/seng-468---copy/seng_468 - Copy/frontend
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/frontend:latest .
cd -

echo 'building and pushing image log-service:latest'
cd source/seng-468---copy/seng_468 - Copy/log_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/log-service:latest .
cd -

echo 'building and pushing image mongo-db:latest'
cd source/seng-468---copy/seng_468 - Copy/mongodb
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/mongo-db:latest .
cd -

echo 'building and pushing image quote-cache:latest'
cd source/seng-468---copy/seng_468 - Copy/quote_cache
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-cache:latest .
cd -

echo 'building and pushing image quote-server:latest'
cd source/seng-468---copy/seng_468 - Copy/quote_server
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-server:latest .
cd -

echo 'building and pushing image worker:latest'
cd source/seng-468---copy/seng_468 - Copy/worker_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/worker:latest .
cd -

echo 'building and pushing image backend-api'
cd source/seng-468---copy/seng_468 - Copy/backend_api
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/backend-api .
cd -

echo 'building and pushing image log-service'
cd source/seng-468---copy/seng_468 - Copy/log_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/log-service .
cd -

echo 'building and pushing image m'
cd source/seng-468---copy/seng_468 - Copy/cli_app
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/m .
cd -

echo 'building and pushing image queue-service'
cd source/seng-468---copy/seng_468 - Copy/queue_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/queue-service .
cd -

echo 'building and pushing image quote-server'
cd source/seng-468---copy/seng_468 - Copy/quote_server
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-server .
cd -

echo 'building and pushing image redis-cache'
cd source/seng-468---copy/seng_468 - Copy/quote_cache
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/redis-cache .
cd -

echo 'building and pushing image worker-service'
cd source/seng-468---copy/seng_468 - Copy/worker_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile  --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/worker-service .
cd -

echo 'done'
