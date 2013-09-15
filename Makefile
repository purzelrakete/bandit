.PHONY: all build test lint coverage deps clean

LIBS := \
github.com/purzelrakete/bandit \
github.com/purzelrakete/bandit/http

BINS := \
github.com/purzelrakete/bandit/api \
github.com/purzelrakete/bandit/example \
github.com/purzelrakete/bandit/job \
github.com/purzelrakete/bandit/plot

PKGS := $(LIBS) $(BINS)

all: deps build lint test

build:
	go build -v $(LIBS)
	go build -o bandit-api github.com/purzelrakete/bandit/api
	go build -o bandit-example github.com/purzelrakete/bandit/example
	go build -o bandit-job github.com/purzelrakete/bandit/job
	go build -o bandit-plot github.com/purzelrakete/bandit/plot

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

clean:
	go clean $(PKGS)
	find . -type f -perm -o+rx -name 'bandit-*' -delete
	find . -type f -name '*.svg' -delete
