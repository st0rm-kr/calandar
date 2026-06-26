package calendar

import (
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleMonth(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	from, ok := parseTimeParam(c, "from")
	if !ok {
		return
	}
	to, ok := parseTimeParam(c, "to")
	if !ok {
		return
	}

	filter := c.Query("filter")
	if filter == "" {
		filter = FilterAll
	}
	sources := parseSourcesParam(c)

	items, err := h.service.Month(c.Request.Context(), userID, from, to, filter, sources)
	if err != nil {
		logger.Errorf("calendar_month_failed user_id=%s from=%s to=%s filter=%s error=%q", userID, from.Format(time.RFC3339), to.Format(time.RFC3339), filter, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, items)
}

func (h *Handler) HandleSubscriptions(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	subscriptions, err := h.service.Subscriptions(c.Request.Context(), userID)
	if err != nil {
		logger.Errorf("calendar_subscriptions_failed user_id=%s error=%q", userID, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, subscriptions)
}

func parseSourcesParam(c *gin.Context) []string {
	raw, exists := c.GetQuery("sources")
	if !exists {
		return nil
	}
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	sources := make([]string, 0, len(parts))
	for _, part := range parts {
		source := strings.TrimSpace(part)
		if source != "" {
			sources = append(sources, source)
		}
	}
	return sources
}

func parseTimeParam(c *gin.Context, name string) (time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		respondError(c, http.StatusBadRequest, "bad_request", "missing "+name+" timestamp")
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid "+name+" timestamp")
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
