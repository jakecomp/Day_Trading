FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
EXPOSE 8002
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./... 

CMD ["go","run","/go/src/app/quote_server.go"]