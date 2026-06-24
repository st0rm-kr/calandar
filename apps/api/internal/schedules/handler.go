package schedules

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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

type scheduleRequest struct {
	Title      string     `json:"title"`
	StartAt    time.Time  `json:"start_at" binding:"required"`
	EndAt      *time.Time `json:"end_at"`
	Location   *string    `json:"location"`
	Visibility string     `json:"visibility"`
}

func (h *Handler) HandleList(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	from, to, ok := parseRange(c)
	if !ok {
		return
	}

	list, err := h.service.ListByRange(c.Request.Context(), userID, from, to)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, list)
}

func (h *Handler) HandleCreate(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	var req scheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	result, err := h.service.CreateManual(c.Request.Context(), userID, CreateManualInput{
		Title:      req.Title,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Location:   req.Location,
		Visibility: req.Visibility,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof(
		"schedule_created user_id=%s schedule_id=%d visibility=%s conflict_count=%d",
		userID,
		result.Schedule.ID,
		result.Schedule.Visibility,
		len(result.Conflicts),
	)
	respondOK(c, http.StatusCreated, result)
}

func (h *Handler) HandleUpdate(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	var req scheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	result, err := h.service.UpdateManual(c.Request.Context(), userID, scheduleID, UpdateManualInput{
		Title:      req.Title,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Location:   req.Location,
		Visibility: req.Visibility,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof(
		"schedule_updated user_id=%s schedule_id=%d visibility=%s conflict_count=%d",
		userID,
		result.Schedule.ID,
		result.Schedule.Visibility,
		len(result.Conflicts),
	)
	respondOK(c, http.StatusOK, result)
}

func (h *Handler) HandleDelete(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteManual(c.Request.Context(), userID, scheduleID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof("schedule_deleted user_id=%s schedule_id=%d", userID, scheduleID)
	respondOK(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) HandleConflicts(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}

	start, ok := parseTimeParam(c, "start")
	if !ok {
		return
	}
	var end *time.Time
	if raw := c.Query("end"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			respondError(c, http.StatusBadRequest, "bad_request", "invalid end timestamp")
			return
		}
		utc := parsed.UTC()
		end = &utc
	}
	var excludeID *int64
	if raw := c.Query("exclude_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			respondError(c, http.StatusBadRequest, "bad_request", "invalid exclude_id")
			return
		}
		excludeID = &parsed
	}

	conflicts, err := h.service.FindConflicts(c.Request.Context(), userID, start, end, excludeID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, conflicts)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidScheduleTitle),
		errors.Is(err, ErrInvalidScheduleTime),
		errors.Is(err, ErrInvalidVisibility):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrCannotModifyEventSchedule):
		respondError(c, http.StatusConflict, "event_schedule", err.Error())
	case errors.Is(err, ErrScheduleNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "not_found", "schedule not found")
	default:
		logger.Errorf("schedule_service_error path=%s error=%q", c.Request.URL.Path, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func parseRange(c *gin.Context) (time.Time, time.Time, bool) {
	from, ok := parseTimeParam(c, "from")
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	to, ok := parseTimeParam(c, "to")
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
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

func parseScheduleID(c *gin.Context) (int64, bool) {
	scheduleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || scheduleID <= 0 {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid schedule id")
		return 0, false
	}
	return scheduleID, true
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
