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

@echo off
IF "%1"=="" GOTO DEFAULT_CONTAINER_RUNTIME
SET CONTAINER_RUNTIME=%1%
GOTO :MAIN

:DEFAULT_CONTAINER_RUNTIME
    SET CONTAINER_RUNTIME=docker

:UNSUPPORTED_BUILD_SYSTEM
    echo 'Unsupported build system passed as an argument for pushing the images.'
    GOTO SKIP

:MAIN
IF NOT %CONTAINER_RUNTIME% == "docker" IF NOT %CONTAINER_RUNTIME% == "podman" GOTO UNSUPPORTED_BUILD_SYSTEM
REM go to the parent directory so that all the relative paths will be correct
cd ..

echo "building image backend:latest"
pushd source\seng-468---copy\seng_468 - Copy\backend_api
%CONTAINER_RUNTIME% build -f Dockerfile -t backend:latest .
popd

echo "building image frontend:latest"
pushd source\seng-468---copy\seng_468 - Copy\frontend
%CONTAINER_RUNTIME% build -f Dockerfile -t frontend:latest .
popd

echo "building image log-service:latest"
pushd source\seng-468---copy\seng_468 - Copy\log_service
%CONTAINER_RUNTIME% build -f Dockerfile -t log-service:latest .
popd

echo "building image mongo-db:latest"
pushd source\seng-468---copy\seng_468 - Copy\mongodb
%CONTAINER_RUNTIME% build -f Dockerfile -t mongo-db:latest .
popd

echo "building image quote-cache:latest"
pushd source\seng-468---copy\seng_468 - Copy\quote_cache
%CONTAINER_RUNTIME% build -f Dockerfile -t quote-cache:latest .
popd

echo "building image quote-server:latest"
pushd source\seng-468---copy\seng_468 - Copy\quote_server
%CONTAINER_RUNTIME% build -f Dockerfile -t quote-server:latest .
popd

echo "building image worker:latest"
pushd source\seng-468---copy\seng_468 - Copy\worker_service
%CONTAINER_RUNTIME% build -f Dockerfile -t worker:latest .
popd

echo "building image backend-api"
pushd source\seng-468---copy\seng_468 - Copy\backend_api
%CONTAINER_RUNTIME% build -f Dockerfile -t backend-api .
popd

echo "building image log-service"
pushd source\seng-468---copy\seng_468 - Copy\log_service
%CONTAINER_RUNTIME% build -f Dockerfile -t log-service .
popd

echo "building image m"
pushd source\seng-468---copy\seng_468 - Copy\cli_app
%CONTAINER_RUNTIME% build -f Dockerfile -t m .
popd

echo "building image queue-service"
pushd source\seng-468---copy\seng_468 - Copy\queue_service
%CONTAINER_RUNTIME% build -f Dockerfile -t queue-service .
popd

echo "building image quote-server"
pushd source\seng-468---copy\seng_468 - Copy\quote_server
%CONTAINER_RUNTIME% build -f Dockerfile -t quote-server .
popd

echo "building image redis-cache"
pushd source\seng-468---copy\seng_468 - Copy\quote_cache
%CONTAINER_RUNTIME% build -f Dockerfile -t redis-cache .
popd

echo "building image worker-service"
pushd source\seng-468---copy\seng_468 - Copy\worker_service
%CONTAINER_RUNTIME% build -f Dockerfile -t worker-service .
popd

echo "done"

:SKIP
