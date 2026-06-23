package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-secret"

func TestRequireUserRejectsMissingBearerToken(t *testing.T) {
	router := testRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	const want = `{"data":null,"error":{"code":"unauthorized","message":"missing bearer token"}}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRequireUserRejectsInvalidBearerToken(t *testing.T) {
	router := testRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	const want = `{"data":null,"error":{"code":"unauthorized","message":"invalid bearer token"}}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRequireUserStoresValidTokenSubjectInRequestContext(t *testing.T) {
	router := testRouter()
	sub := "d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, testJWTSecret, map[string]any{
		"sub": sub,
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	}))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	const want = `{"data":{"user_id":"d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"},"error":null}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", RequireUser(testJWTSecret), func(c *gin.Context) {
		userID, ok := UserIDFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"data": nil, "error": "missing user id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"user_id": userID.String()}, "error": nil})
	})
	return router
}

func signHS256JWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
