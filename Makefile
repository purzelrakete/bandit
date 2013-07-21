.PHONY: all plot build test deps clean

all: deps build test plot

plot:
	make -C plot

build:
	go build -v

test:
	go test -v

deps:
	go get -v

clean:
	go clean
