
.PHONY: travis
travis: clean build test

.PHONY: all
all: clean build binary test

.PHONY: build test clean
clean:
	rm -f dmap-server

build:
	echo "${PWD}"
	go build .

binary:
	go build -o dmap-server server/main.go

test:
	go test

run:
	./dmap-server

