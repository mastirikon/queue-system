FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем готовый бинарник с хоста
COPY bin/api-linux ./api

EXPOSE 8080

CMD ["./api"]
