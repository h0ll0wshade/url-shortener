package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/h0ll0wshade/url-shortener/internal/model"
	"github.com/h0ll0wshade/url-shortener/internal/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}
// Resgister User steps
// 	 validate — is email valid? is password 8+ chars?
//         ↓
//   check MongoDB — does this email already exist?
//         ↓
//   hash the password using bcrypt
//         ↓
//   save User to MongoDB
//         ↓
//   create a JWT with { user_id, email, exp }
func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, string, error) {
	// check if email already exists
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("email already exists")
	}

	// hash the password — never store plain text
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// build the user
	user := &model.User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		PasswordHash: string(hash),
	}

	// save to MongoDB
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// create JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 
// find user in MongoDB by email
// ↓
// compare password against stored hash using bcrypt
// ↓
// if match → create JWT

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	// find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	// compare password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// generate JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// generateJWT — creates a signed JWT for the user
func (s *AuthService) generateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}