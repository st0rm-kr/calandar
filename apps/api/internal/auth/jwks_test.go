package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestVerifierAcceptsES256TokenFromJWKS(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	const kid = "test-kid"

	jwks := jwksResponse{Keys: []jwksKey{{
		Kid: kid,
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
		Y:   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
	}}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	sub := "d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": sub,
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	verifier := NewVerifier("", server.URL)
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.GET("/protected", verifier.Middleware(), func(c *gin.Context) {
		userID, ok := UserIDFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"data": nil, "error": "missing user id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"user_id": userID.String()}, "error": nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), sub) {
		t.Fatalf("expected subject %s in body, got %s", sub, rec.Body.String())
	}
}

func TestVerifierRejectsES256WhenJWKSDisabled(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": "d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8",
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "any"
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	verifier := NewVerifier(testJWTSecret, "")
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.GET("/protected", verifier.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": nil, "error": nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for ES256 when JWKS disabled, got %d", rec.Code)
	}
}
