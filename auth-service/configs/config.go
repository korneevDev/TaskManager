package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBHost             string        `mapstructure:"DB_HOST"`
	DBPort             string        `mapstructure:"DB_PORT"`
	DBUser             string        `mapstructure:"DB_USER"`
	DBPassword         string        `mapstructure:"DB_PASSWORD"`
	DBName             string        `mapstructure:"DB_NAME"`
	JWTSecret          string        `mapstructure:"JWT_SECRET"`
	AccessTokenExpiry  time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRY"`
	RefreshTokenExpiry time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRY"`
	GRPCPort           string        `mapstructure:"GRPC_PORT"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.AutomaticEnv()

	cfg := &Config{
		DBHost:             viper.GetString("DB_HOST"),
		DBPort:             viper.GetString("DB_PORT"),
		DBUser:             viper.GetString("DB_USER"),
		DBPassword:         viper.GetString("DB_PASSWORD"),
		DBName:             viper.GetString("DB_NAME"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		AccessTokenExpiry:  viper.GetDuration("ACCESS_TOKEN_EXPIRY"),
		RefreshTokenExpiry: viper.GetDuration("REFRESH_TOKEN_EXPIRY"),
		GRPCPort:           viper.GetString("GRPC_PORT"),
	}

	return cfg, nil
}
