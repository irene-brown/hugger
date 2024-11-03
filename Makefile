# Makefile for building and installing the command line tool

# Variables
APP_NAME = hugger
SRC_DIR = cmd-cli
MAIN_FILE = $(SRC_DIR)/main.go
BUILD_DIR = build
LINUX_BIN = $(BUILD_DIR)/$(APP_NAME)_linux
WINDOWS_BIN = $(BUILD_DIR)/$(APP_NAME)_windows.exe
DARWIN_BIN=$(BUILD_DIR)/$(APP_NAME)_darwin

GO=go
FLAGS=-ldflags '-s -w'

# Create build directory if it doesn't exist
.PHONY: all
all: build

# Build for Linux
.PHONY: build-linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(FLAGS) -o $(LINUX_BIN) $(MAIN_FILE)

# Build for Windows
.PHONY: build-windows
build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(FLAGS) -o $(WINDOWS_BIN) $(MAIN_FILE)
# Build for Darwin (Mac OS X)
.PHONY: build-darwin
build-darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(FLAGS) -o $(DARWIN_BIN) $(MAIN_FILE)
# Build for all platforms
.PHONY: build
build: build-linux build-windows build-darwin

# Install the application
.PHONY: install
install: build
	@echo "Installing $(APP_NAME)..."
	@cp $(LINUX_BIN) /usr/local/bin/$(APP_NAME) || true
	@cp $(DARWIN_BIN) /Applications/$(APP_NAME) || true
	@cp $(WINDOWS_BIN) C:\Program Files\$(APP_NAME)\$(APP_NAME).exe || true

# Clean up build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Help message
.PHONY: help
help:
	@echo "Makefile for $(APP_NAME)"
	@echo "Usage:"
	@echo "  make build          Build the application for both Linux and Windows"
	@echo "  make build-linux    Build the application for Linux"
	@echo "  make build-windows  Build the application for Windows"
	@echo "  make build-darwin   Build the application for Mac OS X"
	@echo "  make install        Install the application"
	@echo "  make clean          Clean up build artifacts"
	@echo "  make help           Show this help message"
