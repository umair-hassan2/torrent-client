build:
	go build -o bin/main main.go
run:
	./bin/main
test:
	go test ./... -v