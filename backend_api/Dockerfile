FROM golang:1.19

WORKDIR /app

# Download necessary Go modules
COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /godocker

EXPOSE 8000

CMD [ "/godocker" ]