# Makefile for Booklet Maker Go Application

# Build the application
build:
	go build -o bin/booklet-maker .

# Install the application
install:
	go install .

# Run the application
run:
	go run . -input test.pdf

# Test the application
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Tidy the go modules
tidy:
	go mod tidy

# Build for different platforms
build-all: build-linux build-windows build-mac

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/booklet-maker-linux-amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/booklet-maker-windows-amd64.exe .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/booklet-maker-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/booklet-maker-darwin-arm64 .

.PHONY: build install run test clean fmt vet tidy build-all build-linux build-windows build-mac