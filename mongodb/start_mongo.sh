#!/bin/bash 
docker build -t mongodb . 
docker run -d --name mongodb_cont mongodb