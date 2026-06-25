package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	CookieName string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	return Config{
		AppPort:    getEnv("APP_PORT", "8037"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "Product_Inventory_437"),
		JWTSecret:  getEnv("JWT_SECRET", "change-this-secret"),
		CookieName: getEnv("AUTH_COOKIE_NAME", "Product_Inventory_437_token"),
	}
}

func (c Config) ServerAddr() string {
	return fmt.Sprintf(":%s", c.AppPort)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
