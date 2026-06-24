package calendar

import (
	"net/http"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/auth"
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

	items, err := h.service.Month(c.Request.Context(), userID, from, to, filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, items)
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
