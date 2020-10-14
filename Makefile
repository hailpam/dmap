
.PHONY: build test clean
clean:
	rm -f dmap-server

build:
	go build .
	go build -o dmap-server server/main.go

test:
	go test

run:
	./dmap-server

.PHONY: all
all: clean build test
