package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/h0ll0wshade/url-shortener/internal/service"
)

type URLHandler struct {
	urlService *service.URLService
	baseURL    string
}

func NewURLHandler(urlService *service.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		baseURL:    baseURL,
	}
}

// POST /urls
func (h *URLHandler) CreateAnonymous(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.urlService.CreateAnonymous(c.Request.Context(), req.OriginalURL)
	if err != nil {
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

	// check if link has expired
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