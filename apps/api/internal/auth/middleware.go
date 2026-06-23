package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireUser(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "missing bearer token"}})
			c.Abort()
			return
		}

		tokenText := strings.TrimPrefix(header, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithAudience("authenticated"), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "invalid bearer token"}})
			c.Abort()
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "token subject is missing"}})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "token subject is invalid"}})
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), userIDKey, userID))
		c.Next()
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
