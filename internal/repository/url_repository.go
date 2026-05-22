package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/h0ll0wshade/url-shortener/internal/model"
)

type URLRepository struct {
	collection *mongo.Collection
}

func NewURLRepository(db *mongo.Database) *URLRepository {
	return &URLRepository{
		collection: db.Collection("urls"),
	}
}

// save a new URL to MongoDB
func (r *URLRepository) Create(ctx context.Context, url *model.URL) error {
	_, err := r.collection.InsertOne(ctx, url)
	return err
}

// find a URL by alias
func (r *URLRepository) FindByAlias(ctx context.Context, alias string) (*model.URL, error) {
	var url model.URL
	err := r.collection.FindOne(ctx, bson.M{"_id": alias}).Decode(&url)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil // not found, not an error
	}
	return &url, err
}