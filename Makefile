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
	@echo "  git-status          - View git status with component grouping"
	@echo "  git-log             - View recent commit history"
	@echo "  git-diff            - View staged vs unstaged changes"
	@echo "  git-check           - Run pre-commit validation"
	@echo "  help                - Show this help message"

# Git workflow helpers
.PHONY: git-status
git-status:
	@echo "=== Git Status with Branch Info ==="
	@git status -sb
	@echo "\n=== Uncommitted Files by Component ==="
	@git status --short | grep -E "^ M|^M |^A |^ A" | awk '{print $$2}' | \
		sed -E 's|^(publisher|subscriber|event|config|errors)\.go|\1|; s|^redis/(.*)|\1 (redis)|; s|^events/(.*)|\1 (events)|; s|^examples/(.*)|\1 (examples)|; s|^testutil/(.*)|\1 (testutil)|' | \
		sort -u | sed 's/^/  - /' || echo "  (no uncommitted files)"

.PHONY: git-log
git-log:
	@git log --oneline --graph --decorate -20

.PHONY: git-diff
git-diff:
	@echo "=== Staged Changes ==="
	@git diff --cached --stat
	@echo "\n=== Unstaged Changes ==="
	@git diff --stat

.PHONY: git-check
git-check:
	@echo "=== Pre-Commit Checklist ==="
	@echo "[ ] One component/concern per commit?"
	@echo "[ ] Message follows Conventional Commits?"
	@echo "[ ] Files related to same feature?"
	@echo "[ ] < 10 files, < 500 lines?"
	@echo ""
	@echo "=== Running Tests ==="
	@go test ./... -short
	@echo ""
	@echo "=== Running Linter ==="
	@golangci-lint run --timeout=5m
