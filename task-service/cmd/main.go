package main

import (
	"log"

	_ "github.com/korneevDev/task-service/docs"

	"github.com/gin-gonic/gin"
	configs "github.com/korneevDev/task-service/configs"
	"github.com/korneevDev/task-service/internal/handlers"
	"github.com/korneevDev/task-service/internal/models"
	"github.com/korneevDev/task-service/internal/repository"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Task Service API
// @version 1.0
func main() {
	cfg, err := configs.LoadConfig()
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
	db.AutoMigrate(&models.Task{})

	taskRepo := repository.NewTaskRepository(db)
	taskHandler := handlers.NewTaskHandler(*taskRepo)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	taskGroup := r.Group("/tasks")
	//  taskGroup.Use(authMiddleware())
	{
		taskGroup.POST("", taskHandler.Create)
		taskGroup.GET("", taskHandler.List)
		// taskGroup.GET("/:id", taskHandler.Get)
		// taskGroup.PUT("/:id", taskHandler.Update)
		// taskGroup.DELETE("/:id", taskHandler.Delete)
	}

	r.Run(":8081")
}
