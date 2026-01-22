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
	@docker compose build

docker-up: ## Запустить все сервисы в Docker
	@docker compose up -d

docker-down: ## Остановить все сервисы
	@docker compose down

docker-logs: ## Посмотреть логи
	@docker compose logs -f

docker-monitor: ## Открыть Asynq Web UI
	@echo "Opening Asynq Monitor at http://localhost:8081"
	@open http://localhost:8081 2>/dev/null || xdg-open http://localhost:8081 2>/dev/null || echo "Open http://localhost:8081 in your browser"

build-linux: ## Собрать бинарники для Linux (vdska)
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-linux ./cmd/api
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/worker-linux ./cmd/worker
	@echo "Done! Binaries: bin/api-linux, bin/worker-linux"

deploy: build-linux ## Собрать и задеплоить на vdska
	@echo "Uploading to vdska..."
	@scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
	@scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/
	@echo "Restarting services on vdska..."
	@ssh root@vdska "cd /home/finance-system/queue-system && docker compose up -d --build"
	@echo "✅ Deployed successfully!"
	@echo "Check logs: ssh root@vdska 'docker compose logs -f'"

deploy-full: build-linux ## Задеплоить всё (включая конфиги)
	@echo "Uploading everything to vdska..."
	@scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
	@scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/
	@scp docker-compose.yml root@vdska:/home/finance-system/queue-system/
	@scp .env.production root@vdska:/home/finance-system/queue-system/.env
	@echo "Restarting services on vdska..."
	@ssh root@vdska "cd /home/finance-system/queue-system && docker compose up -d --build"
	@echo "✅ Deployed successfully!"

logs-remote: ## Посмотреть логи на vdska
	@ssh root@vdska "cd /home/finance-system/queue-system && docker compose logs -f"

status-remote: ## Проверить статус на vdska
	@ssh root@vdska "cd /home/finance-system/queue-system && docker compose ps"

test: ## Запустить тесты
	@go test -v ./...

clean: ## Очистить бинарники
	@rm -rf bin/
	@echo "Cleaned!"

.DEFAULT_GOAL := help