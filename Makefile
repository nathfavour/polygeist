.PHONY: build install test

INSTALL_DIR ?= $(HOME)/.local/bin
LDFLAGS = -s -w

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/polygeist ./cmd/polygeist

install:
	./install.sh

test:
	go test ./...
