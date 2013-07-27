.PHONY: all build test deps install clean

PKGS := \
github.com/purzelrakete/bandit \
github.com/purzelrakete/bandit/http \
github.com/purzelrakete/bandit/oob \
github.com/purzelrakete/bandit/plot

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
