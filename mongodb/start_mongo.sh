#!/bin/bash 
docker build -t mongodb . 
docker run -d -p 27017:27017 --network day_trading_network --name mongodb_cont mongodb