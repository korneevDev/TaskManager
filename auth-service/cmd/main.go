package main

import (
	"log"

	"github.com/gin-gonic/gin"
	config "github.com/korneevDev/auth-service/configs"
	"github.com/korneevDev/auth-service/internal/handlers"
	"github.com/korneevDev/auth-service/internal/models"
	"github.com/korneevDev/auth-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключаемся к БД
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
	authHandler := handlers.NewAuthHandler(*userRepo)

	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)

	r.Run(":8080")
}
