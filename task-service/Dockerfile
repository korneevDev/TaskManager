# Используем многостадийную сборку
# Стадия сборки
FROM golang:1.21-alpine as builder

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /task-service ./cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из стадии сборки
COPY --from=builder /task-service /app/task-service
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml

# Копируем миграции (если используете)
COPY migrations /app/migrations

# Указываем порт, который будет использоваться
EXPOSE 8081

# Команда для запуска приложения
CMD ["/app/task-service"]