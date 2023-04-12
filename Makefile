PACKAGE=sonic
BINARY=sonic
VERSION=1.0.0
DATE=date+%FT%T%z
COMMIT = $(shell git rev-parse --short HEAD)

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildDate=${DATE} -X main.commit=$(COMMIT)"

.PHONY: all
all: clean build

build:
	@echo "编译本地版本"
	GO111MODULE=on go build -mod=vendor ${LDFLAGS} -o $(BINARY)
	@file ${BINARY}

build-linux:
	@echo "编译linux版本"
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor ${LDFLAGS} -o $(BINARY)
	@file ${BINARY}

install:
	go install ${LDFLAGS}

clean:
	rm -rf logs
	go clean
