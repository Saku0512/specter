SHELL := /bin/bash

APP := specter
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)
CONFIG ?= config.yml
PORT ?= 8080
GOCACHE := $(CURDIR)/.cache/go-build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build version init run try test clean

build:
	@mkdir -p "$(BIN_DIR)" "$(GOCACHE)"
	GOCACHE="$(GOCACHE)" go build -ldflags "-X main.version=$(VERSION)" -o "$(BIN)" .

version: build
	./"$(BIN)" --version

init: build
	./"$(BIN)" init -o "$(CONFIG)"

run: build
	./"$(BIN)" -c "$(CONFIG)" -p "$(PORT)"

try: version

test:
	@mkdir -p "$(GOCACHE)"
	GOCACHE="$(GOCACHE)" go test ./...

clean:
	rm -rf "$(BIN_DIR)" "$(CURDIR)/.cache"
