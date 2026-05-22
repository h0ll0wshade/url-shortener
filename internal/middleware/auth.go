package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth — rejects requests without a valid JWT
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := parseToken(c, jwtSecret)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuth — extracts user_id if token present, otherwise continues anonymously
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := parseToken(c, jwtSecret)
		c.Set("user_id", userID) // may be "" if no token or invalid
		c.Next()
	}
}

// parseToken — shared helper, returns user_id or empty string
func parseToken(c *gin.Context, jwtSecret string) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
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