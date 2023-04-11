:: Copyright IBM Corporation 2021
::
::  Licensed under the Apache License, Version 2.0 (the "License");
::   you may not use this file except in compliance with the License.
::   You may obtain a copy of the License at
::
::        http://www.apache.org/licenses/LICENSE-2.0
::
::  Unless required by applicable law or agreed to in writing, software
::  distributed under the License is distributed on an "AS IS" BASIS,
::  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
::  See the License for the specific language governing permissions and
::  limitations under the License.

for /F "delims=" %%i in ("%cd%") do set basename="%%~ni"

if not %basename% == "scripts" (
    echo "please run this script from the 'scripts' directory"
    exit 1
)

REM go to the parent directory so that all the relative paths will be correct
cd ..

@echo off
IF "%3"=="" GOTO DEFAULT_PLATFORMS
SET PLATFORMS=%3%
GOTO :REGISTRY

:DEFAULT_PLATFORMS
    SET PLATFORMS=linux/amd64,linux/arm64,linux/s390x,linux/ppc64le

:REGISTRY
    IF "%2"=="" GOTO DEFAULT_REGISTRY
    IF "%1"=="" GOTO DEFAULT_REGISTRY
    SET REGISTRY_URL=%1
    SET REGISTRY_NAMESPACE=%2
    GOTO :MAIN

:DEFAULT_REGISTRY
    SET REGISTRY_URL=quay.io
    SET REGISTRY_NAMESPACE=stockapp

:MAIN
:: Uncomment the below line if you want to enable login before pushing
:: docker login %REGISTRY_URL%

echo "building and pushing image backend:latest"
pushd source\seng-468---copy\seng_468 - Copy\backend_api
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/backend:latest .
popd

echo "building and pushing image frontend:latest"
pushd source\seng-468---copy\seng_468 - Copy\frontend
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/frontend:latest .
popd

echo "building and pushing image log-service:latest"
pushd source\seng-468---copy\seng_468 - Copy\log_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/log-service:latest .
popd

echo "building and pushing image mongo-db:latest"
pushd source\seng-468---copy\seng_468 - Copy\mongodb
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/mongo-db:latest .
popd

echo "building and pushing image quote-cache:latest"
pushd source\seng-468---copy\seng_468 - Copy\quote_cache
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-cache:latest .
popd

echo "building and pushing image quote-server:latest"
pushd source\seng-468---copy\seng_468 - Copy\quote_server
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-server:latest .
popd

echo "building and pushing image worker:latest"
pushd source\seng-468---copy\seng_468 - Copy\worker_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/worker:latest .
popd

echo "building and pushing image backend-api"
pushd source\seng-468---copy\seng_468 - Copy\backend_api
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/backend-api .
popd

echo "building and pushing image log-service"
pushd source\seng-468---copy\seng_468 - Copy\log_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/log-service .
popd

echo "building and pushing image m"
pushd source\seng-468---copy\seng_468 - Copy\cli_app
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/m .
popd

echo "building and pushing image queue-service"
pushd source\seng-468---copy\seng_468 - Copy\queue_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/queue-service .
popd

echo "building and pushing image quote-server"
pushd source\seng-468---copy\seng_468 - Copy\quote_server
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/quote-server .
popd

echo "building and pushing image redis-cache"
pushd source\seng-468---copy\seng_468 - Copy\quote_cache
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/redis-cache .
popd

echo "building and pushing image worker-service"
pushd source\seng-468---copy\seng_468 - Copy\worker_service
docker buildx build --platform ${PLATFORMS} -f Dockerfile --push --tag ${REGISTRY_URL}/${REGISTRY_NAMESPACE}/worker-service .
popd

echo "done"
