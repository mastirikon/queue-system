#!/bin/bash

echo "🔨 Building for Linux..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-linux ./cmd/api
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/worker-linux ./cmd/worker

echo "📦 Uploading to vdska..."
scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/

echo "🚀 Restarting services..."
ssh root@vdska "cd /home/finance-system/queue-system && docker compose -f docker-compose-simple.yml up -d --build"

echo "✅ Done! Check logs:"
echo "ssh root@vdska 'cd /home/finance-system/queue-system && docker compose -f docker-compose-simple.yml logs -f'"