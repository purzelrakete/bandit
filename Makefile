.PHONY: all build test deps install clean

PKGS := $(shell echo "github.com/purzelrakete/bandit{,/http,/oob,/plot}")

all: install

build: deps
	go build -v $(PKGS)

test: deps
	go test -v

deps:
	go get -v

install: test
	go install -v $(PKGS)

clean:
	go clean $(PKGS)
