package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    Port       string
    MongoURI   string
    MongoDB    string
    RedisAddr  string
    JWTSecret  string
    BaseURL    string
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, reading from environment")
    }

    return &Config{
        Port:      getEnv("PORT", "8080"),
        MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
        MongoDB:   getEnv("MONGO_DB", "urlshortener"),
        RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
        JWTSecret: getEnv("JWT_SECRET", "changeme"),
        BaseURL:   getEnv("BASE_URL", "http://localhost:8080"),
    }
}

func getEnv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}