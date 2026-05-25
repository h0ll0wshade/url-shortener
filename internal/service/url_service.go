package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/h0ll0wshade/url-shortener/internal/model"
	"github.com/h0ll0wshade/url-shortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const aliasLength = 6

type URLService struct {
	urlRepo *repository.URLRepository
}

func NewURLService(urlRepo *repository.URLRepository) *URLService {
	return &URLService{urlRepo: urlRepo}
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
	return s.urlRepo.FindByAlias(ctx, alias)
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