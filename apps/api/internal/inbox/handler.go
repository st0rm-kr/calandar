package inbox

import (
	"net/http"

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
		logger.Errorf("inbox_list_failed user_id=%s error=%q", userID, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, result)
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
