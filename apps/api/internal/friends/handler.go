package friends

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/logger"
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
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	friends, err := h.service.ListFriends(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, friends)
}

func (h *Handler) HandleListRequests(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	requests, err := h.service.ListRequests(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	respondOK(c, http.StatusOK, requests)
}

func (h *Handler) HandleSendRequest(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	var req struct {
		AddresseeID string `json:"addressee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	addresseeID, err := uuid.Parse(req.AddresseeID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid addressee_id")
		return
	}
	friendship, err := h.service.SendRequest(c.Request.Context(), userID, addresseeID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof("friend_request_sent requester_id=%s addressee_id=%s request_id=%d status=%s", userID, addresseeID, friendship.ID, friendship.Status)
	respondOK(c, http.StatusCreated, friendship)
}

func (h *Handler) HandleAccept(c *gin.Context) {
	h.act(c, h.service.Accept)
}

func (h *Handler) HandleReject(c *gin.Context) {
	h.act(c, h.service.Reject)
}

func (h *Handler) act(c *gin.Context, fn func(ctx context.Context, actorID uuid.UUID, requestID int64) (Friendship, error)) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	requestID, ok := parseID(c)
	if !ok {
		return
	}
	friendship, err := fn(c.Request.Context(), userID, requestID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	logger.Infof("friend_request_transition actor_id=%s request_id=%d status=%s", userID, requestID, friendship.Status)
	respondOK(c, http.StatusOK, friendship)
}

func (h *Handler) HandleDelete(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
		return
	}
	friendID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid user id")
		return
	}
	if err := h.service.DeleteFriend(c.Request.Context(), userID, friendID); err != nil {
		logger.Errorf("friend_delete_failed actor_id=%s friend_id=%s error=%q", userID, friendID, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	logger.Infof("friend_deleted actor_id=%s friend_id=%s", userID, friendID)
	respondOK(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCannotFriendSelf):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrAlreadyFriends):
		respondError(c, http.StatusConflict, "already_friends", err.Error())
	case errors.Is(err, ErrNotAuthorized):
		respondError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, ErrRequestNotFound):
		respondError(c, http.StatusNotFound, "not_found", "friend request not found")
	default:
		logger.Errorf("friend_service_error path=%s error=%q", c.Request.URL.Path, err)
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "bad_request", "invalid request id")
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
