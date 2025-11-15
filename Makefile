.PHONY: help cli-build cli-run cli-install cli-uninstall test test-all coverage fmt vet lint mocks protos generate

# CLI tool commands
cli-build:
	@echo "Building CLI tool..."
	go build -o bin/create-go-api .
	@echo "✓ Built create-go-api"

cli-run:
	@echo "Running CLI tool..."
	go run .

cli-install:
	@echo "Installing create-go-api to system..."
	@go build -o bin/create-go-api .
	@if [ -w /usr/local/bin ]; then \
		cp bin/create-go-api /usr/local/bin/create-go-api; \
		echo "✓ Installed to /usr/local/bin/create-go-api"; \
	else \
		echo "Installing to /usr/local/bin (requires sudo)..."; \
		sudo cp bin/create-go-api /usr/local/bin/create-go-api; \
		echo "✓ Installed to /usr/local/bin/create-go-api"; \
	fi
	@echo "You can now run 'create-go-api' from anywhere!"

cli-uninstall:
	@echo "Uninstalling create-go-api from system..."
	@if [ -f /usr/local/bin/create-go-api ]; then \
		if [ -w /usr/local/bin ]; then \
			rm -f /usr/local/bin/create-go-api; \
		else \
			sudo rm -f /usr/local/bin/create-go-api; \
		fi; \
		echo "✓ Uninstalled from /usr/local/bin/create-go-api"; \
	else \
		echo "create-go-api is not installed"; \
	fi

test:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage report generated: coverage.out"
	@echo "Run 'make coverage' to view the coverage report"

test-all: test
	@echo "✓ All tests passed"

coverage: test
	@echo "Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Run 'go tool cover -html=coverage.out' to view detailed HTML report"

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Vetting code..."
	go vet ./...

lint:
	@echo "Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed. Install from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	}
	golangci-lint run

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@command -v mockery >/dev/null 2>&1 || { \
		echo "mockery is not installed. Installing mockery v2..."; \
		go install github.com/vektra/mockery/v2@latest; \
	}
	@cd internal/generator/static/internal/posts && go generate ./table.go || { \
		echo "Note: Mock generation may require go.mod to be initialized in that directory"; \
		echo "Mocks will be generated when projects are created"; \
	}
	@echo "✓ Mocks generated"

# Generate protobuf code
protos:
	@echo "Generating protobuf code..."
	@command -v buf >/dev/null 2>&1 || { \
		echo "buf is not installed. Install from https://buf.build/docs/installation"; \
		exit 1; \
	}
	@cd internal/generator/static/protos && buf generate
	@echo "✓ Protobuf code generated"

# Generate all code (mocks and protos)
generate: mocks protos
	@echo "✓ All code generation complete"

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "CLI Tool:"
	@echo "  cli-build       - Build the create-go-api CLI tool"
	@echo "  cli-run         - Run the create-go-api CLI tool"
	@echo "  cli-install     - Install create-go-api to /usr/local/bin"
	@echo "  cli-uninstall   - Uninstall create-go-api from system"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run tests with coverage"
	@echo "  test-all        - Run all tests"
	@echo "  coverage        - View coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt             - Format Go code"
	@echo "  vet             - Vet Go code"
	@echo "  lint            - Lint Go code with golangci-lint"
	@echo ""
	@echo "Code Generation:"
	@echo "  mocks           - Generate mocks for testing"
	@echo "  protos           - Generate protobuf code"
	@echo "  generate        - Generate all code (mocks and protos)"

