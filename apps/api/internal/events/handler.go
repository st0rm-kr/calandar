package events

import (
	"errors"
	"net/http"
	"strconv"
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

type rsvpRequest struct {
	RSVP          string `json:"rsvp" binding:"required"`
	AddToCalendar *bool  `json:"add_to_calendar"`
	Visibility    string `json:"visibility"`
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

func (h *Handler) HandleRSVP(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	var req rsvpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	addToCalendar := true
	if req.AddToCalendar != nil {
		addToCalendar = *req.AddToCalendar
	}

	result, err := h.service.RSVP(c.Request.Context(), userID, eventID, RSVPInput{
		RSVP:          req.RSVP,
		AddToCalendar: addToCalendar,
		Visibility:    req.Visibility,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, result)
}

func (h *Handler) HandleDeleteRSVP(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	result, err := h.service.RSVP(c.Request.Context(), userID, eventID, RSVPInput{
		RSVP:          RSVPNotGoing,
		AddToCalendar: false,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, result)
}

func (h *Handler) HandleCancel(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	event, err := h.service.Cancel(c.Request.Context(), userID, eventID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, event)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEventTitle),
		errors.Is(err, ErrInvalidEventType),
		errors.Is(err, ErrInvalidEventScope),
		errors.Is(err, ErrInvalidEventTime),
		errors.Is(err, ErrInvalidCapacity),
		errors.Is(err, ErrInvalidRSVP):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrCapacityFull):
		respondError(c, http.StatusConflict, "capacity_full", err.Error())
	case errors.Is(err, ErrEventUnavailable):
		respondError(c, http.StatusConflict, "event_unavailable", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "not_found", "event not found")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func parseEventID(c *gin.Context) (int64, bool) {
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid event id")
		return 0, false
	}
	return eventID, true
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
