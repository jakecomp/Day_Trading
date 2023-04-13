@echo off

REM This  batch script ensures that the source mount points in exist on the host.

setlocal enableextensions
echo "Ensuring mount points exist..."
md "%USERPROFILE%\.kube"
md "%USERPROFILE%\.minikube"
endlocal