package users

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidDisplayName = errors.New("display name must be 1-40 characters")
	ErrInvalidAvatarURL   = errors.New("avatar_url must be a valid URL")
	ErrInvalidSearchQuery = errors.New("search query must be at least 2 characters")
)

type ProfileRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Profile, error)
	UpdateMe(ctx context.Context, id uuid.UUID, params UpdateMeParams) (Profile, error)
	Search(ctx context.Context, query string) ([]Profile, error)
}

type Service struct {
	repo ProfileRepository
}

func NewService(repo ProfileRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (Profile, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *Service) UpdateMe(ctx context.Context, userID uuid.UUID, params UpdateMeParams) (Profile, error) {
	displayName := strings.TrimSpace(params.DisplayName)
	if displayName == "" || utf8.RuneCountInString(displayName) > 40 {
		return Profile{}, ErrInvalidDisplayName
	}
	if params.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*params.AvatarURL)
		if avatarURL == "" {
			params.AvatarURL = nil
		} else {
			parsed, err := url.ParseRequestURI(avatarURL)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return Profile{}, ErrInvalidAvatarURL
			}
			params.AvatarURL = &avatarURL
		}
	}
	params.DisplayName = displayName
	return s.repo.UpdateMe(ctx, userID, params)
}

func (s *Service) Search(ctx context.Context, query string) ([]Profile, error) {
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < 2 {
		return nil, ErrInvalidSearchQuery
	}
	return s.repo.Search(ctx, query)
}
