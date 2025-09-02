# Makefile for visualizer project

APP_NAME=visualizer

COMPILER=go

# Build directory
BIN_DIR=bin


# Detect current OS
ifeq ($(OS),Windows_NT)
    CURRENT_OS := windows
    EXE_EXT := .exe
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CURRENT_OS := linux
        EXE_EXT :=
    else ifeq ($(UNAME_S),Darwin)
        CURRENT_OS := darwin
        EXE_EXT :=
    else ifeq ($(UNAME_S),FreeBSD)
        CURRENT_OS := freebsd
        EXE_EXT :=
    else ifeq ($(UNAME_S),OpenBSD)
        CURRENT_OS := openbsd
        EXE_EXT :=
    else ifeq ($(UNAME_S),NetBSD)
        CURRENT_OS := netbsd
        EXE_EXT :=
    else
        CURRENT_OS := unknown
        EXE_EXT :=
    endif
endif

# Target OS (default to current OS)
TARGET_OS ?= $(CURRENT_OS)

# Set executable extension based on target OS
ifeq ($(TARGET_OS),windows)
    TARGET_EXT := .exe
else
    TARGET_EXT :=
endif


.PHONY: all clean build run dev deps build-linux build-windows build-darwin

all: build

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

deps:
	go mod tidy
	go mod download

build: deps $(BIN_DIR)
	GOOS=$(CURRENT_OS) GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)$(TARGET_EXT) ./cmd

build-linux: $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ./cmd
	# GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 ./cmd

build-windows: $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd
	# GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe ./cmd

build-darwin: $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 ./cmd

run: build
	./$(BIN_DIR)/$(APP_NAME)$(TARGET_EXT)

dev:
	go run ./cmd

clean:
	rm -rf $(BIN_DIR)

