.DEFAULT_GOAL := up

COMPOSE := docker compose --env-file config.env
TEST_COMPOSE := docker compose -p test-wallet-tests -f docker-compose.test.yaml
LOAD_RATE ?= 1000
LOAD_SECONDS ?= 60

.PHONY: up down db migrate migrate-status logs ps docs test test-integration test-load test-env test-down

# Build and run all
up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

# Run only PG
db:
	$(COMPOSE) up -d --wait postgres

# Run migrations
migrate: db
	$(COMPOSE) run --rm migrator

migrate-status: db
	$(COMPOSE) run --rm migrator status

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps -a

# Regenerate docs
docs:
	go run github.com/swaggo/swag/cmd/swag@latest init \
		-g main.go \
		-d ./cmd/api,./internal/httpapi \
		-o ./docs \
		--parseInternal

# Unit and HTTP tests, including the Go race detector.
test:
	go test -race -cover ./...

# The test project has its own database volume and does not publish ports.
test-env:
	$(TEST_COMPOSE) up -d --wait postgres
	$(TEST_COMPOSE) run --rm migrator

test-integration: test-env
	$(TEST_COMPOSE) run --build --rm --no-deps tests

# One wallet, 1000 deposits per second; verify the final balance.
test-load: test-env
	$(TEST_COMPOSE) up --build -d --wait api
	$(TEST_COMPOSE) run --rm --no-deps -e RATE=$(LOAD_RATE) -e DURATION_SECONDS=$(LOAD_SECONDS) k6

test-down:
	$(TEST_COMPOSE) down
