package events

import (
	"errors"
	"net/http"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
}

type createEventRequest struct {
	Title    string     `json:"title"`
	Type     string     `json:"type"`
	Scope    string     `json:"scope"`
	GroupID  *int64     `json:"group_id"`
	StartAt  time.Time  `json:"start_at" binding:"required"`
	EndAt    *time.Time `json:"end_at"`
	Location *string    `json:"location"`
	Capacity *int       `json:"capacity"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleCreate(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	event, err := h.service.Create(c.Request.Context(), userID, CreateEventInput{
		Title:    req.Title,
		Type:     req.Type,
		Scope:    req.Scope,
		GroupID:  req.GroupID,
		StartAt:  req.StartAt,
		EndAt:    req.EndAt,
		Location: req.Location,
		Capacity: req.Capacity,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	respondOK(c, http.StatusCreated, event)
}

func (h *Handler) HandleGetDetail(c *gin.Context) {
	detail, err := h.service.GetDetail(c.Request.Context(), c.Param("slug"))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, detail)
}

func (h *Handler) HandleListMine(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	events, err := h.service.ListMine(c.Request.Context(), userID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, events)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEventTitle),
		errors.Is(err, ErrInvalidEventType),
		errors.Is(err, ErrInvalidEventScope),
		errors.Is(err, ErrInvalidEventTime),
		errors.Is(err, ErrInvalidCapacity):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "not_found", "event not found")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
