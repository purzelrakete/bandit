.PHONY: all build test lint coverage deps install clean

PKGS := \
github.com/purzelrakete/bandit \
github.com/purzelrakete/bandit/http \
github.com/purzelrakete/bandit/oob \
github.com/purzelrakete/bandit/plot \
github.com/purzelrakete/bandit/example

all: deps build lint test install

build:
	go build -v $(PKGS)

test:
	go test -v $(PKGS)

lint:
	if find . -name '*.go' | xargs golint | grep ":"; then false; else true; fi

coverage:
	goveralls -service drone.io $${COVERALLS_TOKEN:?}

deps:
	go get -v $(PKGS)
	go get github.com/axw/gocov/gocov
	go get github.com/golang/lint/golint
	go get github.com/mattn/goveralls

install:
	go install -v $(PKGS)

clean:
	go clean $(PKGS)
