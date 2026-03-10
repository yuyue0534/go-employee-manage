package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	GinMode     string
}

func loadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] no .env file, reading from environment")
	}
	return &Config{
		DatabaseURL: mustGetEnv("DATABASE_URL"),
		ServerPort:  getEnvOrDefault("SERVER_PORT", "8080"),
		GinMode:     getEnvOrDefault("GIN_MODE", "debug"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("[config] required env var %q is not set", key)
	}
	return v
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
