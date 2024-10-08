# Makefile

# Set environment variables
CGO_ENABLED := 0
GOOS := linux
GOARCH := amd64

# Go file
TARGET := main

# Default target
all: build

build:
	@echo "Building $(TARGET)..."
	go build -o $(TARGET) $(TARGET).go

# Build target
build-linux:
	@echo "Building $(TARGET) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -o $(TARGET) $(TARGET).go

# Clean up generated files
clean:
	@echo "Cleaning up..."
	rm -f $(TARGET)

# Run the program
run:
	@echo "Running $(TARGET)..."
	go build $(TARGET).go