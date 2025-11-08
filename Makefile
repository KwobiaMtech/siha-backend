# Makefile for HealthyPay Backend

# Variables
APP_NAME := siha
DOCKER_COMPOSE := docker compose
DOCKER := docker
GO := go

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

.PHONY: help build run stop clean test dev logs restart status health

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Environment setup
setup: ## Set up the development environment
	@echo "$(YELLOW)Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(GREEN)Created .env file from .env.example$(NC)"; \
		echo "$(YELLOW)Please edit .env file with your configuration$(NC)"; \
	fi

# Build targets
build: ## Build the Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	$(DOCKER_COMPOSE) build --no-cache

build-fast: ## Build the Docker image (with cache)
	@echo "$(YELLOW)Building Docker image (fast)...$(NC)"
	$(DOCKER_COMPOSE) build

build-local: ## Build for local development
	@echo "$(YELLOW)Building for local development...$(NC)"
	$(GO) build -o $(APP_NAME) main.go

docker-build: ## Build Docker image directly
	@echo "Removing existing Docker images..."
	-docker rmi siha:latest 2>/dev/null || true
	-docker images | grep siha | awk '{print $$3}' | xargs -r docker rmi 2>/dev/null || true
	-docker image prune -f 2>/dev/null || true
	@echo "Building Docker image..."
	docker build --no-cache -t siha:latest .

# Development targets
dev: setup ## Start development environment
	@echo "$(YELLOW)Starting development environment...$(NC)"
	$(DOCKER_COMPOSE) up --build -d
	@echo "$(GREEN)Development environment started!$(NC)"
	@echo "$(YELLOW)Application: http://localhost:8090$(NC)"

dev-local: build-local ## Start local development environment
	@echo "$(YELLOW)Starting local development environment...$(NC)"
	@echo "$(YELLOW)Run './$(APP_NAME)' to start the application$(NC)"

# Production targets
up: ## Start the application in production mode
	@echo "$(YELLOW)Starting application...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Application started!$(NC)"

run: build up ## Build and run the application

run-local: build-local ## Build and run locally
	@echo "$(YELLOW)Starting application locally...$(NC)"
	./$(APP_NAME)

# Control targets
stop: ## Stop the application
	@echo "$(YELLOW)Stopping application...$(NC)"
	$(DOCKER_COMPOSE) stop
	@echo "$(GREEN)Application stopped!$(NC)"

down: ## Stop and remove containers
	@echo "$(YELLOW)Stopping and removing containers...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Containers removed!$(NC)"

restart: stop up ## Restart the application

# Monitoring targets
logs: ## Show application logs
	$(DOCKER_COMPOSE) logs -f

logs-app: ## Show only application logs
	$(DOCKER_COMPOSE) logs -f app

status: ## Show container status
	$(DOCKER_COMPOSE) ps

health: ## Check application health
	@echo "$(YELLOW)Checking application health...$(NC)"
	@curl -f http://localhost:8090/health || echo "$(RED)Application is not healthy$(NC)"

# Testing targets
test: ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO) test ./...

test-verbose: ## Run tests with verbose output
	@echo "$(YELLOW)Running tests (verbose)...$(NC)"
	$(GO) test -v ./...

# Maintenance targets
clean: ## Remove containers, images, and volumes
	@echo "$(YELLOW)Cleaning up...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER) system prune -f
	@echo "$(GREEN)Cleanup completed!$(NC)"

clean-all: ## Remove everything including images
	@echo "$(YELLOW)Cleaning up everything...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER) system prune -af
	@echo "$(GREEN)Full cleanup completed!$(NC)"

update: ## Update and restart the application
	@echo "$(YELLOW)Updating application...$(NC)"
	git pull
	$(DOCKER_COMPOSE) build --no-cache
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Application updated!$(NC)"

# Production deployment targets
deploy: ## Deploy to production
	@echo "$(YELLOW)Deploying to production...$(NC)"
	$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml up -d --build
	@echo "$(GREEN)Production deployment completed!$(NC)"

docker-deploy: ## Build Docker image and deploy container
	@$(MAKE) docker-build
	@echo "$(YELLOW)Deploying container...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Docker build and deployment completed!$(NC)"

# Utility targets
shell: ## Access application container shell
	$(DOCKER_COMPOSE) exec app sh

exec: ## Execute command in app container (usage: make exec CMD="command")
	$(DOCKER_COMPOSE) exec app $(CMD)

# Security targets
security-scan: ## Run security scan on Docker image
	@echo "$(YELLOW)Running security scan...$(NC)"
	$(DOCKER) run --rm -v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD):/app \
		aquasec/trivy image $(APP_NAME)_app

# Development convenience targets
format: ## Format Go code
	@echo "$(YELLOW)Formatting Go code...$(NC)"
	$(GO) fmt ./...

lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	golangci-lint run

mod-tidy: ## Tidy Go modules
	@echo "$(YELLOW)Tidying Go modules...$(NC)"
	$(GO) mod tidy

# Quick start target
quickstart: setup build up ## Quick start for new developers
	@echo "$(GREEN)Quick start completed!$(NC)"
	@echo "$(YELLOW)Application running at: http://localhost:8090$(NC)"
	@echo "$(YELLOW)Run 'make logs' to see application logs$(NC)"
	@echo "$(YELLOW)Run 'make help' to see all available commands$(NC)"