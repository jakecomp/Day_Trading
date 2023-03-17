FROM redis:alpine3.17

WORKDIR /usr/src/code  

RUN apk add jq 
RUN apk add curl
EXPOSE 6379 

# ENV REDIS_ARGS="-h 10.9.0.10 -p 6379"

COPY . .
# CMD ["sh","/usr/src/code/update_quotes.sh"]