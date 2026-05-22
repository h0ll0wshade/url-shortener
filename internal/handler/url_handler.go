package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/h0ll0wshade/url-shortener/internal/model"
	"github.com/h0ll0wshade/url-shortener/internal/service"
)

type URLHandler struct {
	urlService *service.URLService
	baseURL    string
	jwtSecret  string
}

func NewURLHandler(urlService *service.URLService, baseURL, jwtSecret string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		baseURL:    baseURL,
		jwtSecret:  jwtSecret,
	}
}

// POST /urls
func (h *URLHandler) Create(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required,url"`
		CustomAlias string `json:"custom_alias"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// extract userID from JWT if provided (optional)
	userID := h.extractUserID(c)

	// custom alias requires login
	if req.CustomAlias != "" && userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required to use a custom alias"})
		return
	}

	// decide which service function to call
	var url *model.URL
	var err error

	if req.CustomAlias != "" {
		url, err = h.urlService.CreateWithCustomAlias(c.Request.Context(), req.OriginalURL, req.CustomAlias, userID)
	} else {
		url, err = h.urlService.CreateWithRandomAlias(c.Request.Context(), req.OriginalURL, userID)
	}

	if err != nil {
		if err.Error() == "alias already taken" {
			c.JSON(http.StatusConflict, gin.H{"error": "alias already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create short URL"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"alias":        url.Alias,
		"original_url": url.OriginalURL,
		"short_url":    h.baseURL + "/r/" + url.Alias,
		"created_at":   url.CreatedAt,
	})
}

// GET /urls/:alias
func (h *URLHandler) GetByAlias(c *gin.Context) {
	alias := c.Param("alias")

	url, err := h.urlService.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	if url == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "this link has expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alias":        url.Alias,
		"original_url": url.OriginalURL,
		"short_url":    h.baseURL + "/r/" + url.Alias,
		"created_at":   url.CreatedAt,
		"expires_at":   url.ExpiresAt,
	})
}

// extractUserID — reads the JWT if present, returns user_id or empty string
func (h *URLHandler) extractUserID(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	uid, _ := claims["user_id"].(string)
	return uid
}