all: app test

app: main.go
	go build -o ./app -v ./...

test: main.go main_test.go
	go test
