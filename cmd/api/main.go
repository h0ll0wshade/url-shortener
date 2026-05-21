package main

import (
    "context"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/h0ll0wshade/url-shortener/config"
)

func main() {
    // Load config
    cfg := config.Load()

    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
    if err != nil {
        log.Fatal("MongoDB connection failed:", err)
    }

    // Ping to verify connection
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("MongoDB ping failed:", err)
    }
    log.Println("✅ Connected to MongoDB")

    db := client.Database(cfg.MongoDB)
    _ = db // will use this in next steps

    // Start Gin
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    log.Println("🚀 Server running on port", cfg.Port)
    r.Run(":" + cfg.Port)
}