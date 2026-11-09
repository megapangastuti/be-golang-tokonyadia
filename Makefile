.PHONY: run build clean swagger test help build-all

# Default target
help:
	@echo "Available targets:"
	@echo "  run          - Run the application in development mode"
	@echo "  build        - Build the application for current OS"
	@echo "  build-all    - Build for all platforms (Linux, Windows, macOS)"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  clean        - Remove built binaries"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help message"

# Run the application
run:
	go run cmd/api/main.go

# Build for current OS
build:
	go build -o tokonyadia-api cmd/api/main.go

# Build for all platforms
build-all:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build -o tokonyadia-api-linux-amd64 cmd/api/main.go
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build -o tokonyadia-api-linux-arm64 cmd/api/main.go
	@echo "Building for Windows AMD64 (64-bit)..."
	GOOS=windows GOARCH=amd64 go build -o tokonyadia-api-windows-amd64.exe cmd/api/main.go
	@echo "Building for Windows 386 (32-bit)..."
	GOOS=windows GOARCH=386 go build -o tokonyadia-api-windows-386.exe cmd/api/main.go
	@echo "Building for Windows ARM64..."
	GOOS=windows GOARCH=arm64 go build -o tokonyadia-api-windows-arm64.exe cmd/api/main.go
	@echo "Building for macOS AMD64 (Intel)..."
	GOOS=darwin GOARCH=amd64 go build -o tokonyadia-api-macos-amd64 cmd/api/main.go
	@echo "Building for macOS ARM64 (M1/M2/M3)..."
	GOOS=darwin GOARCH=arm64 go build -o tokonyadia-api-macos-arm64 cmd/api/main.go
	@echo "Build completed for all platforms!"

# Generate Swagger documentation
swagger:
	~/go/bin/swag init -g cmd/api/main.go -o docs

# Clean built binaries
clean:
	rm -f tokonyadia-api tokonyadia-api-*

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy
