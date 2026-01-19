.PHONY: setup fmt tidy lint test run build clean

# Setup project dependencies
setup:
	@go mod download
	@go mod tidy

# Format code
fmt:
	@gofmt -l -w .
	@go mod tidy

# Tidy dependencies
tidy:
	@go mod tidy -v

# Run linter
lint:
	@golangci-lint run ./...

# Run tests
test:
	@go test -race -shuffle=on -short -failfast ./...

# Run the server
run:
	@go run cmd/server/main.go

# Run with config file
run-config:
	@go run cmd/server/main.go --config=./config/config.yaml

# Build binary
build:
	@go build -o bin/ai-gateway cmd/server/main.go

# Clean build artifacts
clean:
	@rm -rf bin/

# Generate mocks (if using mockgen)
gen:
	@go generate ./...

# Help
help:
	@echo "Available targets:"
	@echo "  setup      - Download and tidy dependencies"
	@echo "  fmt        - Format code"
	@echo "  tidy       - Tidy go.mod"
	@echo "  lint       - Run linter"
	@echo "  test       - Run tests"
	@echo "  run        - Run server with default config"
	@echo "  run-config - Run server with config file"
	@echo "  build      - Build binary"
	@echo "  clean      - Clean build artifacts"
