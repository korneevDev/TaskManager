FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/main.go

FROM alpine:latest
COPY --from=builder /app/auth-service .
COPY --from=builder /app/.env .
EXPOSE 8080 50051
CMD ["./auth-service"]