package users

import (
	"errors"
	"net/http"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleAuthMe(c *gin.Context) {
	h.handleGetMe(c)
}

func (h *Handler) HandleGetMe(c *gin.Context) {
	h.handleGetMe(c)
}

func (h *Handler) HandlePatchMe(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	var req struct {
		DisplayName string  `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	profile, err := h.service.UpdateMe(c.Request.Context(), userID, UpdateMeParams{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	logger.Infof("user_profile_updated user_id=%s", userID)
	respondOK(c, http.StatusOK, profile)
}

func (h *Handler) HandleSearch(c *gin.Context) {
	profiles, err := h.service.Search(c.Request.Context(), c.Query("q"))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, profiles)
}

func (h *Handler) handleGetMe(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	profile, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, profile)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidDisplayName), errors.Is(err, ErrInvalidAvatarURL), errors.Is(err, ErrInvalidSearchQuery):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "not_found", "profile not found")
	default:
		logger.Errorf("user_service_error path=%s error=%q", c.Request.URL.Path, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
