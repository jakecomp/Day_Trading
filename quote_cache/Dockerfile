FROM redis:alpine3.17

WORKDIR /usr/src/code  

RUN apk add jq 
RUN apk add curl
EXPOSE 6379  

# RUN /usr/local/bin/redis-server --daemonize yes

# ENV REDIS_ARGS="-h 10.9.0.10 -p 6379"

COPY . .
CMD ["/bin/sh","-c","sleep 10;sh update_quotes.sh"]
