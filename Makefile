.PHONY: all build test deps install clean

PKGS := \
github.com/purzelrakete/bandit \
github.com/purzelrakete/bandit/http \
github.com/purzelrakete/bandit/oob \
github.com/purzelrakete/bandit/plot

all: deps build test install

build:
	go build -v $(PKGS)

test:
	go test -v

deps:
	go get -v

install:
	go install -v $(PKGS)

clean:
	go clean $(PKGS)
