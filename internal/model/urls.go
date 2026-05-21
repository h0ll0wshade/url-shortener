package model

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type URL struct {
    Alias            string             `bson:"_id"                json:"alias"`
    OriginalURL      string             `bson:"original_url"       json:"original_url"`
    UserID           primitive.ObjectID `bson:"user_id,omitempty"  json:"user_id,omitempty"`
    LinkPasswordHash string             `bson:"link_password_hash" json:"-"`
    CreatedAt        time.Time          `bson:"created_at"         json:"created_at"`
    ExpiresAt        *time.Time         `bson:"expires_at"         json:"expires_at,omitempty"`
}