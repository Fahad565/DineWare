.PHONY: init-go init-node docs test-backend test-frontend up up-build down logs clean

DOCKER_USER := -u $$(id -u):$$(id -g)

# 1. Download Go dependencies using a cached Docker volume
init-go:
	docker run --rm \
		$(DOCKER_USER) \
		-e GOCACHE=/go/.cache \
		-v "$(PWD)/backend":/app \
		-v dineware-go-cache:/go \
		-w /app \
		golang:1.22-alpine go mod tidy

# 2. Install Node packages inside an ephemeral Node 20 container
init-node:
	docker run --rm \
		$(DOCKER_USER) \
		-v "$(PWD)/frontend":/app \
		-w /app \
		node:20-alpine npm install

# 3. Generate Swagger API docs using the cached Go volume
docs:
	docker run --rm \
		$(DOCKER_USER) \
		-e GOCACHE=/go/.cache \
		-v "$(PWD)/backend":/app \
		-v dineware-go-cache:/go \
		-w /app \
		golang:1.22-alpine sh -c "which swag || go install github.com/swaggo/swag/cmd/swag@latest && swag init"

# 4. Run Go unit tests inside Docker
test-backend: docs
	docker run --rm \
		-e GOCACHE=/go/.cache \
		-v "$(PWD)/backend":/app \
		-v dineware-go-cache:/go \
		-w /app \
		golang:1.22-alpine sh -c "apk add --no-cache gcc musl-dev && go test -race -v ./..."

# 5. Run Next.js linter & build check inside Docker
test-frontend:
	docker run --rm \
		$(DOCKER_USER) \
		-v "$(PWD)/frontend":/app \
		-w /app \
		node:20-alpine sh -c "npm run lint && npm run build"

# 6. Build and run your fullstack app via Docker Compose
up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	docker volume rm dineware-go-cache || true
