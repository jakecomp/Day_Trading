#!/bin/bash

while [ true ] 

do
    content=$(curl -sb -H "Accept: application/json" "http://10.9.0.6:8002")  
    stock_name=$(printf '%s\n' "$content" |jq -r '.stock') 
    stock_name=$(printf '%s\n' "$content" |jq -r '.price')
    #stock_price=$(jq -r '.price' <<<"$content") 
    echo $content
    redis-cli -h 10.9.0.10 -p 6379 SET stock_name stock_price
    sleep 5
done