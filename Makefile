.PHONY: all plot http oob build test deps clean

all: deps build test plot http oob

plot:
	make -C plot

http:
	make -C http

oob:
	make -C oob

build:
	go build -v

test:
	go test -v

deps:
	go get -v

clean:
	go clean
