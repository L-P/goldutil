VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=$(shell basename "$(shell pwd)")

all: $(EXEC)

$(EXEC):
	go build ${BUILDFLAGS}

.PHONY: $(EXEC) test lint windows

windows:
	GOOS=windows GOARCH=amd64 go build ${BUILDFLAGS} 

test:
	go test ./...

lint:
	golangci-lint run
