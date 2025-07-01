package main

import (
	"log"
	"net"

	configs "github.com/korneevDev/auth-service/configs"
	grpcInternal "github.com/korneevDev/auth-service/internal/grpc"
	"github.com/korneevDev/auth-service/internal/models"
	"github.com/korneevDev/auth-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "github.com/korneevDev/auth-service/proto/auth"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := configs.LoadConfig()

	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	dsn := "host=" + cfg.DBHost +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" port=" + cfg.DBPort +
		" sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&models.User{})

	userRepo := repository.NewUserRepository(db)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpcInternal.NewAuthServer(
		*userRepo, cfg.JWTSecret, cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry))

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server started on :" + cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
