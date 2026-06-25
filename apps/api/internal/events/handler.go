package events

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type ogPageData struct {
	Title       string
	Description string
	Slug        string
}

var ogPageTemplate = template.Must(template.New("og_page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{ .Title }}</title>
<meta property="og:title" content="{{ .Title }}">
<meta property="og:type" content="website">
<meta property="og:description" content="{{ .Description }}">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="{{ .Title }}">
<meta name="twitter:description" content="{{ .Description }}">
</head>
<body>
<div id="root" data-event-slug="{{ .Slug }}"></div>
</body>
</html>`))

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

	logger.Infof("event_created owner_id=%s event_id=%d scope=%s type=%s", userID, event.ID, event.Scope, event.Type)
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

func (h *Handler) HandleOGPage(c *gin.Context) {
	detail, err := h.service.GetDetail(c.Request.Context(), c.Param("slug"))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	event := detail.Event
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	if err := ogPageTemplate.Execute(c.Writer, ogPageData{
		Title:       event.Title,
		Description: ogPageDescription(event),
		Slug:        event.ShareSlug,
	}); err != nil {
		_ = c.Error(err)
	}
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
	logger.Infof(
		"event_rsvp_updated user_id=%s event_id=%d rsvp=%s add_to_calendar=%t going_count=%d conflict_count=%d",
		userID,
		eventID,
		result.Participant.RSVP,
		addToCalendar,
		result.GoingCount,
		len(result.Conflicts),
	)
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
	logger.Infof("event_rsvp_removed user_id=%s event_id=%d", userID, eventID)
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
	logger.Infof("event_cancelled owner_id=%s event_id=%d", userID, event.ID)
	respondOK(c, http.StatusOK, event)
}

func (h *Handler) HandleInvite(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	var req struct {
		InviteeIDs []string `json:"invitee_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	inviteeIDs := make([]uuid.UUID, 0, len(req.InviteeIDs))
	for _, raw := range req.InviteeIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			respondError(c, http.StatusBadRequest, "bad_request", "invalid invitee id")
			return
		}
		inviteeIDs = append(inviteeIDs, id)
	}

	invited, err := h.service.Invite(c.Request.Context(), userID, eventID, inviteeIDs)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof("event_invites_created actor_id=%s event_id=%d requested_count=%d invited_count=%d", userID, eventID, len(inviteeIDs), len(invited))
	respondOK(c, http.StatusOK, invited)
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
	case errors.Is(err, ErrInviteForbidden):
		respondError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "not_found", "event not found")
	default:
		logger.Errorf("event_service_error path=%s error=%q", c.Request.URL.Path, err)
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

func ogPageDescription(event Event) string {
	description := event.StartAt.UTC().Format("2006-01-02 15:04 MST")
	if event.Location != nil && *event.Location != "" {
		description += " - " + *event.Location
	}
	return description
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
