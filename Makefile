.PHONY: all

all:
	go build -v ./...
	go test -race -v ./...
