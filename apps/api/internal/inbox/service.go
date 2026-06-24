package inbox

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	KindEventInvite   = "event_invite"
	KindFriendRequest = "friend_request"
	KindGroupInvite   = "group_invite"
)

type EventInvite struct {
	ParticipantID int64
	EventID       int64
	Title         string
	StartAt       time.Time
	EndAt         *time.Time
	Location      *string
	Status        string
	ConflictCount int
	CreatedAt     time.Time
}

type FriendRequest struct {
	ID            int64
	RequesterID   uuid.UUID
	RequesterName string
	CreatedAt     time.Time
}

type GroupInvite struct {
	ID          int64
	GroupID     int64
	GroupName   string
	InviterName string
	CreatedAt   time.Time
}

type Repository interface {
	ListInvitedEvents(ctx context.Context, userID uuid.UUID) ([]EventInvite, error)
	ListPendingFriendRequests(ctx context.Context, userID uuid.UUID) ([]FriendRequest, error)
	ListPendingGroupInvites(ctx context.Context, userID uuid.UUID) ([]GroupInvite, error)
}

type InboxCounts struct {
	Total          int `json:"total"`
	EventInvites   int `json:"event_invites"`
	FriendRequests int `json:"friend_requests"`
	GroupInvites   int `json:"group_invites"`
}

type InboxItem struct {
	ID             string    `json:"id"`
	Kind           string    `json:"kind"`
	SourceID       int64     `json:"source_id"`
	EventID        int64     `json:"event_id,omitempty"`
	GroupID        int64     `json:"group_id,omitempty"`
	Title          string    `json:"title"`
	Subtitle       string    `json:"subtitle"`
	CreatedAt      time.Time `json:"created_at"`
	Disabled       bool      `json:"disabled"`
	DisabledReason string    `json:"disabled_reason,omitempty"`
	ConflictCount  int       `json:"conflict_count,omitempty"`
}

type InboxResult struct {
	Counts InboxCounts `json:"counts"`
	Items  []InboxItem `json:"items"`
}

type Service struct {
	repo Repository
	now  func() time.Time
}

type Option func(*Service)

func WithNow(now func() time.Time) Option {
	return func(s *Service) {
		s.now = now
	}
}

func NewService(repo Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) (InboxResult, error) {
	events, err := s.repo.ListInvitedEvents(ctx, userID)
	if err != nil {
		return InboxResult{}, err
	}
	friendRequests, err := s.repo.ListPendingFriendRequests(ctx, userID)
	if err != nil {
		return InboxResult{}, err
	}
	groupInvites, err := s.repo.ListPendingGroupInvites(ctx, userID)
	if err != nil {
		return InboxResult{}, err
	}

	now := s.now()
	items := make([]InboxItem, 0, len(events)+len(friendRequests)+len(groupInvites))

	for _, invite := range events {
		item := InboxItem{
			ID:            fmt.Sprintf("event_invite:%d", invite.ParticipantID),
			Kind:          KindEventInvite,
			SourceID:      invite.ParticipantID,
			EventID:       invite.EventID,
			Title:         invite.Title,
			Subtitle:      eventSubtitle(invite),
			CreatedAt:     invite.CreatedAt,
			ConflictCount: invite.ConflictCount,
		}
		if invite.Status == "cancelled" {
			item.Disabled = true
			item.DisabledReason = "活动已取消"
		} else if invite.Status == "expired" || endOf(invite).Before(now) {
			item.Disabled = true
			item.DisabledReason = "活动已结束"
		}
		items = append(items, item)
	}

	for _, request := range friendRequests {
		items = append(items, InboxItem{
			ID:        fmt.Sprintf("friend_request:%d", request.ID),
			Kind:      KindFriendRequest,
			SourceID:  request.ID,
			Title:     request.RequesterName,
			Subtitle:  "想加你为好友",
			CreatedAt: request.CreatedAt,
		})
	}

	for _, invite := range groupInvites {
		items = append(items, InboxItem{
			ID:        fmt.Sprintf("group_invite:%d", invite.ID),
			Kind:      KindGroupInvite,
			SourceID:  invite.ID,
			GroupID:   invite.GroupID,
			Title:     invite.GroupName,
			Subtitle:  fmt.Sprintf("%s 邀请你加入群组", invite.InviterName),
			CreatedAt: invite.CreatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return InboxResult{
		Counts: InboxCounts{
			Total:          len(items),
			EventInvites:   len(events),
			FriendRequests: len(friendRequests),
			GroupInvites:   len(groupInvites),
		},
		Items: items,
	}, nil
}

func eventSubtitle(invite EventInvite) string {
	subtitle := invite.StartAt.UTC().Format("2006-01-02 15:04 MST")
	if invite.Location != nil && *invite.Location != "" {
		subtitle += " · " + *invite.Location
	}
	return subtitle
}

func endOf(invite EventInvite) time.Time {
	if invite.EndAt != nil {
		return *invite.EndAt
	}
	return invite.StartAt
}
