package notifications

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) HandleList(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	result, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		logger.Errorf("notification_list_failed user_id=%s error=%q", userID, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, result)
}

func (h *Handler) HandleMarkRead(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid notification id")
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(c, http.StatusNotFound, "not_found", "notification not found")
			return
		}
		logger.Errorf("notification_mark_read_failed user_id=%s notification_id=%d error=%q", userID, id, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	logger.Infof("notification_marked_read user_id=%s notification_id=%d", userID, id)
	respondOK(c, http.StatusOK, gin.H{"read": true})
}

func (h *Handler) HandleMarkAllRead(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		logger.Errorf("notification_mark_all_read_failed user_id=%s error=%q", userID, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	logger.Infof("notifications_marked_all_read user_id=%s", userID)
	respondOK(c, http.StatusOK, gin.H{"read": true})
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
