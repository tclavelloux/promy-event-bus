.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@docker-compose up -d redis
	@sleep 2
	@go test -v -race -cover ./...
	@docker-compose down

.PHONY: test-short
test-short:
	@echo "Running short tests (excluding integration)..."
	go test -v -race -cover -short ./...

.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: docker-up
docker-up:
	@echo "Starting Redis..."
	docker-compose up -d
	@echo "Redis is running on localhost:6379"

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down -v

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: tidy
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf vendor/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

.PHONY: example-publisher
example-publisher:
	@echo "Running publisher example..."
	go run examples/publisher/main.go

.PHONY: example-subscriber
example-subscriber:
	@echo "Running subscriber example..."
	go run examples/subscriber/main.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test                 - Run all tests"
	@echo "  test-integration     - Run tests with Redis via Docker"
	@echo "  test-short          - Run tests excluding integration tests"
	@echo "  coverage            - Generate test coverage report"
	@echo "  docker-up           - Start Redis via Docker Compose"
	@echo "  docker-down         - Stop Docker services"
	@echo "  docker-logs         - View Docker logs"
	@echo "  lint                - Run golangci-lint"
	@echo "  fmt                 - Format code with go fmt"
	@echo "  vet                 - Run go vet"
	@echo "  tidy                - Tidy go.mod"
	@echo "  clean               - Clean build artifacts and cache"
	@echo "  example-publisher   - Run publisher example"
	@echo "  example-subscriber  - Run subscriber example"
	@echo "  help                - Show this help message"
