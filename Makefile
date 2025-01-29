.PHONY: build test clean

# Binary name
BINARY_NAME=initia-vanity

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/initia-vanity

# Run tests
test:
	go test -v ./...

# Clean build files
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run all quality checks
check: test
	go vet ./...
	go fmt ./...

# Install the application
install:
	go install ./cmd/initia-vanity

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 ./cmd/initia-vanity
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 ./cmd/initia-vanity
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 ./cmd/initia-vanity
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe ./cmd/initia-vanity