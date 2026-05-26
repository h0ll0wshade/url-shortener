package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/h0ll0wshade/url-shortener/internal/cache"
	"github.com/h0ll0wshade/url-shortener/internal/model"
	"github.com/h0ll0wshade/url-shortener/internal/metrics"
	"github.com/h0ll0wshade/url-shortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const aliasLength = 6
const cacheTTL = 1 * time.Hour

type URLService struct {
	urlRepo *repository.URLRepository
	cache   *cache.Cache
}

func NewURLService(urlRepo *repository.URLRepository, cache *cache.Cache) *URLService {
	return &URLService{
		urlRepo: urlRepo,
		cache:   cache,
	}
}

// CreateWithRandomAlias — generates a random alias, retries if taken
func (s *URLService) CreateWithRandomAlias(ctx context.Context, originalURL string, userID string) (*model.URL, error) {
	alias, err := s.generateUniqueAlias(ctx)
	if err != nil {
		return nil, err
	}
	return s.save(ctx, alias, originalURL, userID)
}

// CreateWithCustomAlias — uses the provided alias, fails if already taken
func (s *URLService) CreateWithCustomAlias(ctx context.Context, originalURL, customAlias, userID string) (*model.URL, error) {
	// check if alias is already taken
	existing, err := s.urlRepo.FindByAlias(ctx, customAlias)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("alias already taken")
	}
	return s.save(ctx, customAlias, originalURL, userID)
}

// GetByAlias — fetch URL metadata by alias
func (s *URLService) GetByAlias(ctx context.Context, alias string) (*model.URL, error) {
	cacheKey := "alias:" + alias

	// 1. try cache first
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var url model.URL
		if jsonErr := json.Unmarshal([]byte(cached), &url); jsonErr == nil {
			metrics.CacheHitsTotal.With(prometheus.Labels{"cache": "alias"}).Inc()
			return &url, nil
		}
		// JSON decode failed — treat as cache miss, refetch from DB
	}

	metrics.CacheMissesTotal.With(prometheus.Labels{"cache": "alias"}).Inc()

	// 2. cache miss — query MongoDB
	url, err := s.urlRepo.FindByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, nil
	}

	// 3. populate cache for next time
	if data, jsonErr := json.Marshal(url); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), cacheTTL)
	}

	return url, nil
}

// save — internal helper, builds the URL and inserts it
func (s *URLService) save(ctx context.Context, alias, originalURL, userID string) (*model.URL, error) {
	url := &model.URL{
		Alias:       alias,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	// attach userID if logged in
	if userID != "" {
		objID, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			url.UserID = objID
		}
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, err
	}
	return url, nil
}

// generateUniqueAlias — keep trying until we find one not in DB
func (s *URLService) generateUniqueAlias(ctx context.Context) (string, error) {
	for {
		alias := randomString(aliasLength)
		existing, err := s.urlRepo.FindByAlias(ctx, alias)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return alias, nil
		}
	}
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GetUserURLs — returns paginated URLs for a user
func (s *URLService) GetUserURLs(ctx context.Context, userID string, page, limit int64) ([]model.URL, int64, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
	return s.urlRepo.FindByUserID(ctx, objID, page, limit)
}