FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем готовый бинарник с хоста
COPY bin/worker-linux ./worker

CMD ["./worker"]
