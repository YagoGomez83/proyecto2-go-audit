package config

import (
	"log"
	"os"
)

type Config struct {
	JWTSecret string
	Port      string
}

func LoadConfig() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET no está definido")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // default
	}

	return &Config{
		JWTSecret: jwtSecret,
		Port:      port,
	}
}
