# Используйте официальный образ Go как базовый
FROM golang:1.21 AS builder

# Установите рабочий каталог в контейнере
WORKDIR /app

# Копируйте go модули и их зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируйте исходный код проекта
COPY . .

# Соберите приложение
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/app/main.go

# Начните новый этап с scratch
# для минимизации размера образа
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируйте бинарный файл из предыдущего этапа
COPY --from=builder /app/server .
COPY --from=builder /app/config/config.yaml ./config/
COPY --from=builder /app/db/migrations ./db/migrations

# Откройте порт, который используется вашим приложением
EXPOSE 8001

# Запустите бинарный файл
CMD ["./server", "config/config.yaml"]
