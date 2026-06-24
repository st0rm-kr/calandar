package notifications

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("notification not found")

type Repository interface {
	Create(ctx context.Context, notification Notification) error
	List(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, userID uuid.UUID, id int64) (bool, error)
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, typ string, payload Payload) error {
	if payload == nil {
		payload = Payload{}
	}
	return s.repo.Create(ctx, Notification{
		UserID:  userID,
		Type:    typ,
		Payload: payload,
	})
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) (ListResult, error) {
	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return ListResult{}, err
	}
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return ListResult{}, err
	}
	if items == nil {
		items = []Notification{}
	}
	return ListResult{UnreadCount: count, Items: items}, nil
}

func (s *Service) MarkRead(ctx context.Context, userID uuid.UUID, id int64) error {
	ok, err := s.repo.MarkRead(ctx, userID, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}
