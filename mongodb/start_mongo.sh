#!/bin/bash 
docker build -t mongodb . 
docker run -d --network day_trading_network --name mongodb_cont mongodb