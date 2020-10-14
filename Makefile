
.PHONY: all
all: clean build test

.PHONY: build test clean
clean:
	rm -f dmap-server

build:
	echo "${PWD}"
	go build .
	go build -o dmap-server server/main.go

test:
	go test

run:
	./dmap-server

