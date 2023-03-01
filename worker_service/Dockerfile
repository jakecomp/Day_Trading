FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
EXPOSE 8003
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./... 

CMD ["go","run","/go/src/app/worker_service.go"]
