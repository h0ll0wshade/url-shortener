package main

import (
    "context"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/h0ll0wshade/url-shortener/config"
	"github.com/h0ll0wshade/url-shortener/internal/handler"
	"github.com/h0ll0wshade/url-shortener/internal/repository"
	"github.com/h0ll0wshade/url-shortener/internal/service"

    "github.com/h0ll0wshade/url-shortener/internal/middleware"
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
    // ── users ──
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
    
    // ── urls ──
	urlRepo     := repository.NewURLRepository(db)
	urlService  := service.NewURLService(urlRepo)
    urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)

	// set up router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

    // anonymous url routes
	urls := r.Group("/urls")
    urls.Use(middleware.OptionalAuth(cfg.JWTSecret))
    {
        urls.POST("", urlHandler.Create)
        urls.GET("/:alias", urlHandler.GetByAlias)
    }
    // redirect route (no auth needed, never)
    r.GET("/r/:alias", urlHandler.Redirect)

    
    log.Println("🚀 Server running on port", cfg.Port)
    r.Run(":" + cfg.Port)
}