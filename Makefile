VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=$(shell basename "$(shell pwd)")

all: $(EXEC)

$(EXEC):
	go build ${BUILDFLAGS} -o goldutil ./cmd

docs/goldutil.1: goldutil.adoc
	asciidoctor --backend manpage "$<" -o "$@"

docs/index.html: goldutil.adoc
	asciidoctor --backend html "$<" -o "$@"

.PHONY: $(EXEC) test lint windows

windows:
	GOOS=windows GOARCH=amd64 go build ${BUILDFLAGS}

test:
	go test ./...

lint:
	golangci-lint run
