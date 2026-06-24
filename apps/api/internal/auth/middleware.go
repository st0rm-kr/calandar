package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

// Verifier validates Supabase access tokens. It supports asymmetric ES256
// tokens via a JWKS endpoint and legacy symmetric HS256 tokens via a shared
// secret. Either mechanism may be configured; both can be enabled at once.
type Verifier struct {
	secret []byte
	jwks   *keySet
}

func NewVerifier(jwtSecret string, jwksURL string) *Verifier {
	v := &Verifier{}
	if jwtSecret != "" {
		v.secret = []byte(jwtSecret)
	}
	if jwksURL != "" {
		v.jwks = newKeySet(jwksURL)
	}
	return v
}

func (v *Verifier) keyFunc(token *jwt.Token) (interface{}, error) {
	switch token.Method.(type) {
	case *jwt.SigningMethodHMAC:
		if v.secret == nil {
			return nil, errors.New("hmac tokens are not accepted")
		}
		return v.secret, nil
	case *jwt.SigningMethodECDSA:
		if v.jwks == nil {
			return nil, errors.New("ecdsa tokens are not accepted")
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token is missing key id")
		}
		return v.jwks.keyByID(kid)
	default:
		return nil, errors.New("unexpected signing method")
	}
}

func (v *Verifier) validMethods() []string {
	methods := make([]string, 0, 2)
	if v.secret != nil {
		methods = append(methods, "HS256")
	}
	if v.jwks != nil {
		methods = append(methods, "ES256")
	}
	return methods
}

func (v *Verifier) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "missing bearer token"}})
			c.Abort()
			return
		}

		tokenText := strings.TrimPrefix(header, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenText, claims, v.keyFunc,
			jwt.WithValidMethods(v.validMethods()),
			jwt.WithAudience("authenticated"),
			jwt.WithExpirationRequired())
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

// RequireUser builds an HS256-only middleware. Retained for callers and tests
// that rely on shared-secret verification.
func RequireUser(jwtSecret string) gin.HandlerFunc {
	return NewVerifier(jwtSecret, "").Middleware()
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
