package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/h0ll0wshade/url-shortener/internal/model"
	"github.com/h0ll0wshade/url-shortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const aliasLength = 6

type URLService struct {
	urlRepo *repository.URLRepository
}

func NewURLService(urlRepo *repository.URLRepository) *URLService {
	return &URLService{
		urlRepo: urlRepo,
	}
}

// CreateAnonymous — shorten a URL for an anonymous user
func (s *URLService) CreateAnonymous(ctx context.Context, originalURL string) (*model.URL, error) {
	// generate a unique alias
	alias, err := s.generateUniqueAlias(ctx)
	if err != nil {
		return nil, err
	}

	url := &model.URL{
		Alias:       alias,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, err
	}

	return url, nil
}

// GetByAlias — fetch URL metadata by alias
func (s *URLService) GetByAlias(ctx context.Context, alias string) (*model.URL, error) {
	return s.urlRepo.FindByAlias(ctx, alias)
}

// generateUniqueAlias — keep generating until we find one not in DB
func (s *URLService) generateUniqueAlias(ctx context.Context) (string, error) {
	for {
		alias := randomString(aliasLength)

		// check if alias already exists in DB
		existing, err := s.urlRepo.FindByAlias(ctx, alias)
		if err != nil {
			return "", err
		}

		// alias is free — use it
		if existing == nil {
			return alias, nil
		}
		// alias taken — loop and try again
	}
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}