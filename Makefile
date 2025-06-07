# Makefile for building Go application

APP_NAME=cricut_finder
OUTPUT_DIR=bin
GIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LD_FLAGS="-X main.GitSHA=$(GIT_SHA)"

WINDOWS_BINARY=$(OUTPUT_DIR)/$(APP_NAME).exe
LINUX_BINARY=$(OUTPUT_DIR)/$(APP_NAME)

.PHONY: all build clean build-windows docker-build docker-run

all: build

build:
	go build -ldflags=$(LD_FLAGS) -o $(LINUX_BINARY) .

build-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(WINDOWS_BINARY) .

clean:
	rm -rf $(OUTPUT_DIR)
