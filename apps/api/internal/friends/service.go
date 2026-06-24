package friends

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCannotFriendSelf = errors.New("cannot send a friend request to yourself")
	ErrAlreadyFriends   = errors.New("already friends")
	ErrRequestNotFound  = errors.New("friend request not found")
	ErrNotAuthorized    = errors.New("not authorized to act on this request")
)

type Repository interface {
	FindBetween(ctx context.Context, a, b uuid.UUID) (Friendship, bool, error)
	Create(ctx context.Context, friendship Friendship) (Friendship, error)
	UpdateStatus(ctx context.Context, id int64, status string) (Friendship, error)
	FindByID(ctx context.Context, id int64) (Friendship, error)
	ListAccepted(ctx context.Context, userID uuid.UUID) ([]Friend, error)
	ListRequests(ctx context.Context, userID uuid.UUID) ([]Friendship, error)
	DeleteBetween(ctx context.Context, a, b uuid.UUID) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) (Friendship, error) {
	if requesterID == addresseeID {
		return Friendship{}, ErrCannotFriendSelf
	}

	existing, found, err := s.repo.FindBetween(ctx, requesterID, addresseeID)
	if err != nil {
		return Friendship{}, err
	}
	if found {
		switch existing.Status {
		case StatusAccepted:
			return Friendship{}, ErrAlreadyFriends
		case StatusPending:
			return existing, nil
		case StatusRejected:
			return s.repo.UpdateStatus(ctx, existing.ID, StatusPending)
		}
	}

	return s.repo.Create(ctx, Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      StatusPending,
	})
}

func (s *Service) Accept(ctx context.Context, actorID uuid.UUID, requestID int64) (Friendship, error) {
	return s.transition(ctx, actorID, requestID, StatusAccepted)
}

func (s *Service) Reject(ctx context.Context, actorID uuid.UUID, requestID int64) (Friendship, error) {
	return s.transition(ctx, actorID, requestID, StatusRejected)
}

func (s *Service) transition(ctx context.Context, actorID uuid.UUID, requestID int64, status string) (Friendship, error) {
	request, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return Friendship{}, err
	}
	if request.AddresseeID != actorID {
		return Friendship{}, ErrNotAuthorized
	}
	if request.Status != StatusPending {
		return Friendship{}, ErrRequestNotFound
	}
	return s.repo.UpdateStatus(ctx, request.ID, status)
}

func (s *Service) ListFriends(ctx context.Context, userID uuid.UUID) ([]Friend, error) {
	friends, err := s.repo.ListAccepted(ctx, userID)
	if err != nil {
		return nil, err
	}
	if friends == nil {
		return []Friend{}, nil
	}
	return friends, nil
}

func (s *Service) ListRequests(ctx context.Context, userID uuid.UUID) ([]RequestView, error) {
	rows, err := s.repo.ListRequests(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]RequestView, 0, len(rows))
	for _, row := range rows {
		direction := "incoming"
		if row.RequesterID == userID {
			direction = "outgoing"
		}
		views = append(views, RequestView{
			ID:          row.ID,
			RequesterID: row.RequesterID,
			AddresseeID: row.AddresseeID,
			Status:      row.Status,
			Direction:   direction,
		})
	}
	return views, nil
}

func (s *Service) DeleteFriend(ctx context.Context, actorID, friendID uuid.UUID) error {
	return s.repo.DeleteBetween(ctx, actorID, friendID)
}

func (s *Service) AreFriends(ctx context.Context, a, b uuid.UUID) (bool, error) {
	existing, found, err := s.repo.FindBetween(ctx, a, b)
	if err != nil {
		return false, err
	}
	return found && existing.Status == StatusAccepted, nil
}
