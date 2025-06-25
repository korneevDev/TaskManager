package main

import (
	"auth-service/internal/config"
	"auth-service/internal/controller"
	"auth-service/internal/grpc"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Загрузка конфигов
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключение к PostgreSQL
	db, err := repository.NewPostgresDB(cfg.DBURL)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// Инициализация слоёв
	repo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(repo, cfg.JWTSecret)
	authController := controller.NewAuthController(authService)

	// HTTP-сервер (Gin)
	r := gin.Default()
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)

	go func() {
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// gRPC-сервер
	grpcServer := grpc.NewServer()
	authGRPC := grpc.NewAuthServer(cfg.JWTSecret)
	proto.RegisterAuthServiceServer(grpcServer, authGRPC)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("gRPC server failed:", err)
	}
	log.Println("gRPC server started on :50051")
	grpcServer.Serve(lis)
}
