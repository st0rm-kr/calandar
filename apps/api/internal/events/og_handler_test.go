package events

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestOGPage(t *testing.T) {
	startAt := time.Date(2026, 6, 23, 10, 30, 0, 0, time.UTC)
	location := `Bouldering & "Cafe" <script>alert(1)</script>`
	event := Event{
		ID:        42,
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Scope:     "personal",
		Title:     `Climb & Brunch <Friends>`,
		Type:      "climbing",
		StartAt:   startAt,
		Location:  &location,
		Status:    StatusActive,
		ShareSlug: "abc123def4",
	}
	repo := newFakeEventRepository()
	repo.events[event.ID] = event
	repo.eventsBySlug[event.ShareSlug] = event.ID

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(NewService(repo))
	router.GET("/e/:slug", handler.HandleOGPage)
	req := httptest.NewRequest(http.MethodGet, "/e/abc123def4", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	assertContains(t, body, `<meta property="og:title"`)
	assertContains(t, body, `<meta property="og:type" content="website"`)
	assertContains(t, body, `<meta name="twitter:card" content="summary_large_image"`)
	assertContains(t, body, `Climb &amp; Brunch &lt;Friends&gt;`)
	assertContains(t, body, `2026-06-23 10:30 UTC`)
	assertContains(t, body, `Bouldering &amp; &#34;Cafe&#34; &lt;script&gt;alert(1)&lt;/script&gt;`)
	assertContains(t, body, `<div id="root"`)

	if strings.Contains(body, event.Title) || strings.Contains(body, location) {
		t.Fatalf("expected title and location to be HTML-escaped, got body: %s", body)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected body to contain %q; body: %s", needle, haystack)
	}
}
