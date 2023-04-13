#!/bin/bash

# This shell script ensures that the source mount points exist on the host.

echo "Ensuring mount points exist..."
mkdir -p "$HOME/.kube" "$HOME/.minikube"
