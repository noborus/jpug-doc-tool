BINARY_NAME := jpug-doc-tool
SRCS := $(shell git ls-files '*.go')
LDFLAGS := "-X github.com/noborus/jpug-doc-tool/jpugdoc.Version=$(shell git describe --tags --abbrev=0 --always)"

all: build

test: $(SRCS)
	go test ./...

deps:
	go mod tidy

build: deps $(BINARY_NAME)

$(BINARY_NAME): $(SRCS)
	go build -ldflags $(LDFLAGS)

install:
	go install -ldflags $(LDFLAGS)

sys-install: build
	sudo install $(BINARY_NAME) /usr/local/bin

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test deps build install clean
