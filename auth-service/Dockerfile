# Стадия сборки
FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml

EXPOSE 8080

CMD ["/app/auth-service"]