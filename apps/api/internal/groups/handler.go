package groups

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleList(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	list, err := h.service.ListGroups(c.Request.Context(), userID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, list)
}

func (h *Handler) HandleCreate(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	group, err := h.service.CreateGroup(c.Request.Context(), userID, CreateGroupInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusCreated, group)
}

func (h *Handler) HandleGetDetail(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	groupID, ok := parseID(c, "id")
	if !ok {
		return
	}
	detail, err := h.service.GetDetail(c.Request.Context(), userID, groupID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, detail)
}

func (h *Handler) HandleUpdate(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	groupID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	group, err := h.service.UpdateGroup(c.Request.Context(), userID, groupID, UpdateGroupInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, group)
}

func (h *Handler) HandleDissolve(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	groupID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.service.DissolveGroup(c.Request.Context(), userID, groupID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, gin.H{"dissolved": true})
}

func (h *Handler) HandleJoin(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	var req struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	group, err := h.service.JoinByInviteCode(c.Request.Context(), userID, req.InviteCode)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, group)
}

func (h *Handler) HandleInvite(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	var req struct {
		GroupID   int64  `json:"group_id" binding:"required"`
		InviteeID string `json:"invitee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	inviteeID, err := uuid.Parse(req.InviteeID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid invitee_id")
		return
	}
	invite, err := h.service.Invite(c.Request.Context(), userID, req.GroupID, inviteeID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusCreated, invite)
}

func (h *Handler) HandleAcceptInvite(c *gin.Context) {
	h.actOnInvite(c, h.service.AcceptInvite)
}

func (h *Handler) HandleRejectInvite(c *gin.Context) {
	h.actOnInvite(c, h.service.RejectInvite)
}

func (h *Handler) actOnInvite(c *gin.Context, fn func(ctx context.Context, actorID uuid.UUID, inviteID int64) error) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	inviteID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := fn(c.Request.Context(), userID, inviteID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) HandleLeave(c *gin.Context) {
	userID, ok := h.actor(c)
	if !ok {
		return
	}
	groupID, ok := parseID(c, "id")
	if !ok {
		return
	}
	targetID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid user id")
		return
	}
	if targetID != userID {
		respondError(c, http.StatusForbidden, "forbidden", "can only remove yourself")
		return
	}
	if err := h.service.LeaveGroup(c.Request.Context(), userID, groupID); err != nil {
		h.respondServiceError(c, err)
		return
	}
	respondOK(c, http.StatusOK, gin.H{"left": true})
}

func (h *Handler) actor(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return uuid.UUID{}, false
	}
	return userID, true
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidGroupName):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrNotOwner):
		respondError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrNotMember):
		respondError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrNotAuthorized):
		respondError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrOwnerMustTransfer):
		respondError(c, http.StatusConflict, "owner_must_transfer", err.Error())
	case errors.Is(err, ErrGroupDissolved):
		respondError(c, http.StatusConflict, "group_dissolved", err.Error())
	case errors.Is(err, ErrGroupNotFound):
		respondError(c, http.StatusNotFound, "not_found", "group not found")
	case errors.Is(err, ErrInviteNotFound):
		respondError(c, http.StatusNotFound, "not_found", "group invite not found")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func parseID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid id")
		return 0, false
	}
	return id, true
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": gin.H{"code": code, "message": message}})
}
