# Makefile for Scrum Master

.PHONY: build clean test help

# Build the application
build:
	@echo "🔨 Building scrum-master..."
	go build -o bin/scrum-master cmd/scrum-master/main.go
	@echo "✅ Build complete!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Clean complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -cover ./...

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	@echo "✅ Dependencies installed!"

# Build for multiple platforms
build-all: clean
	@echo "🔨 Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o bin/scrum-master-linux cmd/scrum-master/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/scrum-master-mac cmd/scrum-master/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/scrum-master-windows.exe cmd/scrum-master/main.go
	@echo "✅ Multi-platform build complete!"

# Run the application with mock data
run-mock:
	@echo "🚀 Running with mock data..."
	./bin/scrum-master process project-desc.md

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  deps         - Install dependencies"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  run-mock     - Run with mock data"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  help         - Show this help"
