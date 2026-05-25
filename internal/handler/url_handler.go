package handler

import (
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/h0ll0wshade/url-shortener/internal/model"
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
func (h *URLHandler) Create(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required,url"`
		CustomAlias string `json:"custom_alias"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// user_id is set by the OptionalAuth middleware (may be "")
	userID := c.GetString("user_id")

	// custom alias requires login
	if req.CustomAlias != "" && userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required to use a custom alias"})
		return
	}

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

// GET /r/:alias
func (h *URLHandler) Redirect(c *gin.Context) {
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

	// perform the redirect
	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

// GET /users/:userId/urls
func (h *URLHandler) GetUserURLs(c *gin.Context) {
	// userID from the JWT (set by RequireAuth middleware)
	tokenUserID := c.GetString("user_id")
	pathUserID := c.Param("userId")

	// users can only see their own URLs
	if tokenUserID != pathUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you can only see your own urls"})
		return
	}

	// parse query params with defaults
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	urls, total, err := h.urlService.GetUserURLs(c.Request.Context(), tokenUserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	// build response — add short_url to each url
	urlList := make([]gin.H, 0, len(urls))
	for _, url := range urls {
		urlList = append(urlList, gin.H{
			"alias":        url.Alias,
			"original_url": url.OriginalURL,
			"short_url":    h.baseURL + "/r/" + url.Alias,
			"created_at":   url.CreatedAt,
			"expires_at":   url.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"urls":  urlList,
	})
}