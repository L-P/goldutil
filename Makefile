VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=$(shell basename "$(shell pwd)")

all: $(EXEC) docs/index.html

$(EXEC):
	go build ${BUILDFLAGS}

docs/index.html:
	pandoc -f man -o docs/index.html goldutil.1

.PHONY: $(EXEC) test lint windows

windows:
	GOOS=windows GOARCH=amd64 go build ${BUILDFLAGS} 

test:
	go test ./...

lint:
	golangci-lint run
