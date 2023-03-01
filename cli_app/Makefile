all: app test

app: main.go
	go build -o ./app -v ./...

run: app
	./app < input.txt

test: main.go main_test.go
	go test
