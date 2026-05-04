VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
CIFLAGS=-mod=readonly
EXEC=$(shell basename "$(shell pwd)")

all: $(EXEC)

$(EXEC):
	go build ${BUILDFLAGS} -o goldutil ./cmd

docs/goldutil.1: goldutil.adoc
	asciidoctor --backend manpage "$<" -o "$@"

docs/index.html: goldutil.adoc
	asciidoctor --backend html "$<" -o "$@"

.PHONY: $(EXEC) test lint ci-windows ci-linux ci-test

ci-linux:
	GOOS=linux GOARCH=amd64 go build ${CIFLAGS} ${BUILDFLAGS} -o goldutil ./cmd

ci-windows:
	GOOS=windows GOARCH=amd64 go build ${CIFLAGS} ${BUILDFLAGS} -o goldutil.exe ./cmd

test:
	go test ./...

ci-test:
	go test ${CIFLAGS} ./...

lint:
	golangci-lint run
