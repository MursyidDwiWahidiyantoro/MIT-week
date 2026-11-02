# AMBA Docker & Dev Workflow
# ponytail: simple targets wrap docker compose; upgrade if build matrix expands
.PHONY: up dev down build logs ps restart clean test

# Default: start production-like demo stack (backend + nginx frontend)
up:
	docker compose up --build

# Start dev mode with hot-reload (Go Air + SvelteKit Vite)
dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Stop all containers
down:
	docker compose down

# Stop containers and remove volumes (clean cache)
clean:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down -v

# Rebuild all images
build:
	docker compose build

# Follow logs
logs:
	docker compose logs -f

# Container status
ps:
	docker compose ps

# Restart containers
restart:
	docker compose restart

# Run all test suites locally
test:
	cd backend && go test -race ./...
	cd frontend && npm test
