:: Copyright IBM Corporation 2021
::
::  Licensed under the Apache License, Version 2.0 (the "License");
::  you may not use this file except in compliance with the License.
::  You may obtain a copy of the License at
::
::        http://www.apache.org/licenses/LICENSE-2.0
::
::  Unless required by applicable law or agreed to in writing, software
::  distributed under the License is distributed on an "AS IS" BASIS,
::  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
::  See the License for the specific language governing permissions and
::  limitations under the License.

:: Invoke as pushimages.bat <registry_url> <registry_namespace> <container_runtime>

@echo off
IF "%3"=="" GOTO DEFAULT_CONTAINER_RUNTIME
SET CONTAINER_RUNTIME=%3%
GOTO :REGISTRY

:DEFAULT_CONTAINER_RUNTIME
    SET CONTAINER_RUNTIME=docker

:REGISTRY
    IF "%2"=="" GOTO DEFAULT_REGISTRY
    IF "%1"=="" GOTO DEFAULT_REGISTRY
    SET REGISTRY_URL=%1
    SET REGISTRY_NAMESPACE=%2
    GOTO :MAIN

:DEFAULT_REGISTRY
    SET REGISTRY_URL=quay.io
    SET REGISTRY_NAMESPACE=stockapp

:UNSUPPORTED_BUILD_SYSTEM
    echo 'Unsupported build system passed as an argument for pushing the images.'
    GOTO SKIP

:MAIN
IF NOT %CONTAINER_RUNTIME% == "docker" IF NOT %CONTAINER_RUNTIME% == "podman" GOTO UNSUPPORTED_BUILD_SYSTEM
:: Uncomment the below line if you want to enable login before pushing
:: %CONTAINER_RUNTIME% login %REGISTRY_URL%

echo "pushing image worker-service"
%CONTAINER_RUNTIME% tag worker-service %REGISTRY_URL%/%REGISTRY_NAMESPACE%/worker-service
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/worker-service

echo "pushing image backend:latest"
%CONTAINER_RUNTIME% tag backend:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/backend:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/backend:latest

echo "pushing image redis-cache"
%CONTAINER_RUNTIME% tag redis-cache %REGISTRY_URL%/%REGISTRY_NAMESPACE%/redis-cache
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/redis-cache

echo "pushing image quote-server"
%CONTAINER_RUNTIME% tag quote-server %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-server
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-server

echo "pushing image queue-service"
%CONTAINER_RUNTIME% tag queue-service %REGISTRY_URL%/%REGISTRY_NAMESPACE%/queue-service
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/queue-service

echo "pushing image m"
%CONTAINER_RUNTIME% tag m %REGISTRY_URL%/%REGISTRY_NAMESPACE%/m
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/m

echo "pushing image log-service"
%CONTAINER_RUNTIME% tag log-service %REGISTRY_URL%/%REGISTRY_NAMESPACE%/log-service
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/log-service

echo "pushing image backend-api"
%CONTAINER_RUNTIME% tag backend-api %REGISTRY_URL%/%REGISTRY_NAMESPACE%/backend-api
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/backend-api

echo "pushing image worker:latest"
%CONTAINER_RUNTIME% tag worker:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/worker:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/worker:latest

echo "pushing image quote-server:latest"
%CONTAINER_RUNTIME% tag quote-server:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-server:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-server:latest

echo "pushing image quote-cache:latest"
%CONTAINER_RUNTIME% tag quote-cache:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-cache:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/quote-cache:latest

echo "pushing image mongo-db:latest"
%CONTAINER_RUNTIME% tag mongo-db:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/mongo-db:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/mongo-db:latest

echo "pushing image log-service:latest"
%CONTAINER_RUNTIME% tag log-service:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/log-service:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/log-service:latest

echo "pushing image frontend:latest"
%CONTAINER_RUNTIME% tag frontend:latest %REGISTRY_URL%/%REGISTRY_NAMESPACE%/frontend:latest
%CONTAINER_RUNTIME% push %REGISTRY_URL%/%REGISTRY_NAMESPACE%/frontend:latest

echo "done"

:SKIP
