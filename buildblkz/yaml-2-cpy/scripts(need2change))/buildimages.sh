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
CONTAINER_RUNTIME=docker
if [ "$#" -eq 1 ]; then
    CONTAINER_RUNTIME=$1
fi
if [ "${CONTAINER_RUNTIME}" != "docker" ] && [ "${CONTAINER_RUNTIME}" != "podman" ]; then
   echo 'Unsupported container runtime passed as an argument for building the images: '"${CONTAINER_RUNTIME}"
   exit 1
fi
cd .. # go to the parent directory so that all the relative paths will be correct

echo 'building image backend:latest'
cd source/seng-468---copy/seng_468--Copy/backend_api
${CONTAINER_RUNTIME} build -f Dockerfile -t backend:latest .
cd -

echo 'building image frontend:latest'
cd source/seng-468---copy/seng_468--Copy/frontend
${CONTAINER_RUNTIME} build -f Dockerfile -t frontend:latest .
cd -

echo 'building image log-service:latest'
cd source/seng-468---copy/seng_468--Copy/log_service
${CONTAINER_RUNTIME} build -f Dockerfile -t log-service:latest .
cd -

echo 'building image mongo-db:latest'
cd source/seng-468---copy/seng_468--Copy/mongodb
${CONTAINER_RUNTIME} build -f Dockerfile -t mongo-db:latest .
cd -

echo 'building image quote-cache:latest'
cd source/seng-468---copy/seng_468--Copy/quote_cache
${CONTAINER_RUNTIME} build -f Dockerfile -t quote-cache:latest .
cd -

echo 'building image quote-server:latest'
cd source/seng-468---copy/seng_468--Copy/quote_server
${CONTAINER_RUNTIME} build -f Dockerfile -t quote-server:latest .
cd -

echo 'building image worker:latest'
cd source/seng-468---copy/seng_468--Copy/worker_service
${CONTAINER_RUNTIME} build -f Dockerfile -t worker:latest .
cd -

echo 'building image backend-api'
cd source/seng-468---copy/seng_468--Copy/backend_api
${CONTAINER_RUNTIME} build -f Dockerfile -t backend-api .
cd -

echo 'building image log-service'
cd source/seng-468---copy/seng_468--Copy/log_service
${CONTAINER_RUNTIME} build -f Dockerfile -t log-service .
cd -

echo 'building image m'
cd source/seng-468---copy/seng_468--Copy/cli_app
${CONTAINER_RUNTIME} build -f Dockerfile -t m .
cd -

echo 'building image queue-service'
cd source/seng-468---copy/seng_468--Copy/queue_service
${CONTAINER_RUNTIME} build -f Dockerfile -t queue-service .
cd -

echo 'building image quote-server'
cd source/seng-468---copy/seng_468--Copy/quote_server
${CONTAINER_RUNTIME} build -f Dockerfile -t quote-server .
cd -

echo 'building image redis-cache'
cd source/seng-468---copy/seng_468--Copy/quote_cache
${CONTAINER_RUNTIME} build -f Dockerfile -t redis-cache .
cd -

echo 'building image worker-service'
cd source/seng-468---copy/seng_468--Copy/worker_service
${CONTAINER_RUNTIME} build -f Dockerfile -t worker-service .
cd -

echo 'done'
