package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

// FindByUserID — returns paginated URLs owned by a user + total count
func (r *URLRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int64) ([]model.URL, int64, error) {
	filter := bson.M{"user_id": userID}

	// count total documents matching the filter
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// pagination — skip (page-1)*limit, then take limit
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1}) // newest first

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var urls []model.URL
	if err := cursor.All(ctx, &urls); err != nil {
		return nil, 0, err
	}

	return urls, total, nil
}