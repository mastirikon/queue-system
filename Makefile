PHONY: help build run-api run-worker docker-build docker-up docker-down test clean

help: ## Показать помощь
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать бинарники
	@echo "Building API..."
	@go build -o bin/api ./cmd/api
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker
	@echo "Done!"

run-api: ## Запустить API локально
	@go run ./cmd/api

run-worker: ## Запустить Worker локально
	@go run ./cmd/worker

docker-build: ## Собрать Docker образы
	@docker-compose build

docker-up: ## Запустить все сервисы в Docker
	@docker-compose up -d

docker-down: ## Остановить все сервисы
	@docker-compose down

docker-logs: ## Посмотреть логи
	@docker-compose logs -f

docker-monitor: ## Открыть Asynq Web UI
	@echo "Opening Asynq Monitor at http://localhost:8081"
	@open http://localhost:8081 2>/dev/null || xdg-open http://localhost:8081 2>/dev/null || echo "Open http://localhost:8081 in your browser"

test: ## Запустить тесты
	@go test -v ./...

clean: ## Очистить бинарники
	@rm -rf bin/
	@echo "Cleaned!"

.DEFAULT_GOAL := help