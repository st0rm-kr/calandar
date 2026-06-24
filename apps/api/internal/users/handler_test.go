package users

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret"

func TestHandleAuthMeReturnsProfileForTokenSubject(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	email := "li@example.com"
	repo := newFakeRepository([]Profile{{
		ID:          userID,
		DisplayName: "Li",
		Email:       &email,
		Status:      "active",
	}})
	router := testRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, userID))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	body := decodeResponse[Profile](t, rec)
	if body.Data.ID != userID || body.Data.DisplayName != "Li" || body.Data.Email == nil || *body.Data.Email != email {
		t.Fatalf("unexpected profile: %+v", body.Data)
	}
}

func TestHandlePatchMeUpdatesDisplayNameAndAvatarURL(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo := newFakeRepository([]Profile{{
		ID:          userID,
		DisplayName: "Li",
		Status:      "active",
	}})
	router := testRouter(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/me", strings.NewReader(`{"display_name":"Lily","avatar_url":"https://example.com/avatar.png"}`))
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, userID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	body := decodeResponse[Profile](t, rec)
	if body.Data.DisplayName != "Lily" {
		t.Fatalf("expected display_name to be updated, got %q", body.Data.DisplayName)
	}
	if body.Data.AvatarURL == nil || *body.Data.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("expected avatar_url to be updated, got %+v", body.Data.AvatarURL)
	}
}

func TestHandleSearchExcludesDeactivatedUsers(t *testing.T) {
	currentUserID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	activeUserID := uuid.MustParse("a3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	deactivatedUserID := uuid.MustParse("b3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo := newFakeRepository([]Profile{
		{ID: currentUserID, DisplayName: "Current", Status: "active"},
		{ID: activeUserID, DisplayName: "Li Ming", Status: "active"},
		{ID: deactivatedUserID, DisplayName: "Lina", Status: "deactivated"},
	})
	router := testRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/users/search?q=li", nil)
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, currentUserID))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	body := decodeResponse[[]Profile](t, rec)
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 active search result, got %d: %+v", len(body.Data), body.Data)
	}
	if body.Data[0].ID != activeUserID {
		t.Fatalf("expected active user %s, got %+v", activeUserID, body.Data[0])
	}
}

type responseBody[T any] struct {
	Data  T       `json:"data"`
	Error *string `json:"error"`
}

func decodeResponse[T any](t *testing.T, rec *httptest.ResponseRecorder) responseBody[T] {
	t.Helper()

	var body responseBody[T]
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, rec.Body.String())
	}
	return body
}

func testRouter(repo ProfileRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(NewService(repo))
	api := router.Group("/api", auth.RequireUser(testJWTSecret))
	api.GET("/auth/me", handler.HandleAuthMe)
	api.GET("/users/me", handler.HandleGetMe)
	api.PATCH("/users/me", handler.HandlePatchMe)
	api.GET("/users/search", handler.HandleSearch)
	return router
}

func signHS256JWT(t *testing.T, sub uuid.UUID) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(map[string]any{
		"sub": sub.String(),
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, []byte(testJWTSecret))
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

type fakeRepository struct {
	profiles map[uuid.UUID]Profile
}

func newFakeRepository(profiles []Profile) *fakeRepository {
	repo := &fakeRepository{profiles: map[uuid.UUID]Profile{}}
	for _, profile := range profiles {
		repo.profiles[profile.ID] = profile
	}
	return repo
}

func (r *fakeRepository) FindByID(ctx context.Context, id uuid.UUID) (Profile, error) {
	profile, ok := r.profiles[id]
	if !ok {
		return Profile{}, nil
	}
	return profile, nil
}

func (r *fakeRepository) UpdateMe(ctx context.Context, id uuid.UUID, params UpdateMeParams) (Profile, error) {
	profile := r.profiles[id]
	profile.DisplayName = params.DisplayName
	profile.AvatarURL = params.AvatarURL
	r.profiles[id] = profile
	return profile, nil
}

func (r *fakeRepository) Search(ctx context.Context, query string) ([]Profile, error) {
	query = strings.ToLower(query)
	var matches []Profile
	for _, profile := range r.profiles {
		email := ""
		if profile.Email != nil {
			email = *profile.Email
		}
		if profile.Status == "active" &&
			(strings.Contains(strings.ToLower(profile.DisplayName), query) || strings.Contains(strings.ToLower(email), query)) {
			matches = append(matches, profile)
		}
	}
	return matches, nil
}
