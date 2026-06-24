package groups

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidGroupName  = errors.New("name must be 1-60 characters")
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupDissolved    = errors.New("group is dissolved")
	ErrInviteNotFound    = errors.New("group invite not found")
	ErrNotOwner          = errors.New("only the group owner can perform this action")
	ErrNotMember         = errors.New("only group members can perform this action")
	ErrNotAuthorized     = errors.New("not authorized to act on this invite")
	ErrOwnerMustTransfer = errors.New("owner must transfer ownership before leaving a non-empty group")
)

type Repository interface {
	Transaction(ctx context.Context, fn func(Repository) error) error
	CreateGroup(ctx context.Context, group Group) (Group, error)
	FindGroupByID(ctx context.Context, id int64) (Group, error)
	FindGroupByInviteCode(ctx context.Context, code string) (Group, error)
	UpdateGroup(ctx context.Context, id int64, fields map[string]any) (Group, error)
	ListGroupsByMember(ctx context.Context, userID uuid.UUID) ([]Group, error)
	FindMember(ctx context.Context, groupID int64, userID uuid.UUID) (GroupMember, bool, error)
	AddMember(ctx context.Context, member GroupMember) (GroupMember, error)
	RemoveMember(ctx context.Context, groupID int64, userID uuid.UUID) error
	CountMembers(ctx context.Context, groupID int64) (int, error)
	ListMembers(ctx context.Context, groupID int64) ([]Member, error)
	CreateInvite(ctx context.Context, invite GroupInvite) (GroupInvite, error)
	FindInviteByID(ctx context.Context, id int64) (GroupInvite, error)
	FindInvite(ctx context.Context, groupID int64, inviteeID uuid.UUID) (GroupInvite, bool, error)
	UpdateInviteStatus(ctx context.Context, id int64, status string) (GroupInvite, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGroup(ctx context.Context, ownerID uuid.UUID, input CreateGroupInput) (Group, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || utf8.RuneCountInString(name) > 60 {
		return Group{}, ErrInvalidGroupName
	}
	description := normalizeDescription(input.Description)

	var created Group
	err := s.repo.Transaction(ctx, func(repo Repository) error {
		code, err := generateInviteCode()
		if err != nil {
			return err
		}
		created, err = repo.CreateGroup(ctx, Group{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			InviteCode:  code,
			Status:      StatusActive,
		})
		if err != nil {
			return err
		}
		_, err = repo.AddMember(ctx, GroupMember{
			GroupID: created.ID,
			UserID:  ownerID,
			Role:    RoleOwner,
		})
		return err
	})
	if err != nil {
		return Group{}, err
	}
	return created, nil
}

func (s *Service) ListGroups(ctx context.Context, userID uuid.UUID) ([]Group, error) {
	groups, err := s.repo.ListGroupsByMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	if groups == nil {
		return []Group{}, nil
	}
	return groups, nil
}

func (s *Service) IsMember(ctx context.Context, groupID int64, userID uuid.UUID) (bool, error) {
	_, ok, err := s.repo.FindMember(ctx, groupID, userID)
	return ok, err
}

func (s *Service) GetDetail(ctx context.Context, actorID uuid.UUID, groupID int64) (GroupDetail, error) {
	group, err := s.repo.FindGroupByID(ctx, groupID)
	if err != nil {
		return GroupDetail{}, err
	}
	if _, ok, err := s.repo.FindMember(ctx, groupID, actorID); err != nil {
		return GroupDetail{}, err
	} else if !ok {
		return GroupDetail{}, ErrNotMember
	}
	members, err := s.repo.ListMembers(ctx, groupID)
	if err != nil {
		return GroupDetail{}, err
	}
	if members == nil {
		members = []Member{}
	}
	return GroupDetail{Group: group, Members: members}, nil
}

func (s *Service) UpdateGroup(ctx context.Context, actorID uuid.UUID, groupID int64, input UpdateGroupInput) (Group, error) {
	group, err := s.repo.FindGroupByID(ctx, groupID)
	if err != nil {
		return Group{}, err
	}
	if group.OwnerID != actorID {
		return Group{}, ErrNotOwner
	}

	fields := map[string]any{}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" || utf8.RuneCountInString(name) > 60 {
			return Group{}, ErrInvalidGroupName
		}
		fields["name"] = name
	}
	if input.Description != nil {
		desc := normalizeDescription(input.Description)
		if desc == nil {
			fields["description"] = nil
		} else {
			fields["description"] = *desc
		}
	}
	if len(fields) == 0 {
		return group, nil
	}
	return s.repo.UpdateGroup(ctx, groupID, fields)
}

func (s *Service) DissolveGroup(ctx context.Context, actorID uuid.UUID, groupID int64) error {
	group, err := s.repo.FindGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OwnerID != actorID {
		return ErrNotOwner
	}
	return s.repo.Transaction(ctx, func(repo Repository) error {
		if _, err := repo.UpdateGroup(ctx, groupID, map[string]any{"status": StatusDissolved}); err != nil {
			return err
		}
		members, err := repo.ListMembers(ctx, groupID)
		if err != nil {
			return err
		}
		for _, m := range members {
			if err := repo.RemoveMember(ctx, groupID, m.UserID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) JoinByInviteCode(ctx context.Context, userID uuid.UUID, code string) (Group, error) {
	var joined Group
	err := s.repo.Transaction(ctx, func(repo Repository) error {
		group, err := repo.FindGroupByInviteCode(ctx, strings.TrimSpace(code))
		if err != nil {
			return err
		}
		if group.Status != StatusActive {
			return ErrGroupDissolved
		}
		joined = group
		if _, ok, err := repo.FindMember(ctx, group.ID, userID); err != nil {
			return err
		} else if ok {
			return nil
		}
		_, err = repo.AddMember(ctx, GroupMember{
			GroupID: group.ID,
			UserID:  userID,
			Role:    RoleMember,
		})
		return err
	})
	if err != nil {
		return Group{}, err
	}
	return joined, nil
}

func (s *Service) Invite(ctx context.Context, actorID uuid.UUID, groupID int64, inviteeID uuid.UUID) (GroupInvite, error) {
	group, err := s.repo.FindGroupByID(ctx, groupID)
	if err != nil {
		return GroupInvite{}, err
	}
	if group.Status != StatusActive {
		return GroupInvite{}, ErrGroupDissolved
	}
	if _, ok, err := s.repo.FindMember(ctx, groupID, actorID); err != nil {
		return GroupInvite{}, err
	} else if !ok {
		return GroupInvite{}, ErrNotMember
	}

	if existing, ok, err := s.repo.FindInvite(ctx, groupID, inviteeID); err != nil {
		return GroupInvite{}, err
	} else if ok {
		switch existing.Status {
		case InviteStatusPending, InviteStatusAccepted:
			return existing, nil
		case InviteStatusRejected:
			return s.repo.UpdateInviteStatus(ctx, existing.ID, InviteStatusPending)
		}
	}

	return s.repo.CreateInvite(ctx, GroupInvite{
		GroupID:   groupID,
		InviterID: actorID,
		InviteeID: inviteeID,
		Status:    InviteStatusPending,
	})
}

func (s *Service) AcceptInvite(ctx context.Context, actorID uuid.UUID, inviteID int64) error {
	return s.repo.Transaction(ctx, func(repo Repository) error {
		invite, err := repo.FindInviteByID(ctx, inviteID)
		if err != nil {
			return err
		}
		if invite.InviteeID != actorID {
			return ErrNotAuthorized
		}
		if invite.Status != InviteStatusPending {
			return ErrInviteNotFound
		}
		group, err := repo.FindGroupByID(ctx, invite.GroupID)
		if err != nil {
			return err
		}
		if group.Status != StatusActive {
			return ErrGroupDissolved
		}
		if _, ok, err := repo.FindMember(ctx, invite.GroupID, actorID); err != nil {
			return err
		} else if !ok {
			if _, err := repo.AddMember(ctx, GroupMember{
				GroupID: invite.GroupID,
				UserID:  actorID,
				Role:    RoleMember,
			}); err != nil {
				return err
			}
		}
		_, err = repo.UpdateInviteStatus(ctx, inviteID, InviteStatusAccepted)
		return err
	})
}

func (s *Service) RejectInvite(ctx context.Context, actorID uuid.UUID, inviteID int64) error {
	invite, err := s.repo.FindInviteByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.InviteeID != actorID {
		return ErrNotAuthorized
	}
	if invite.Status != InviteStatusPending {
		return ErrInviteNotFound
	}
	_, err = s.repo.UpdateInviteStatus(ctx, inviteID, InviteStatusRejected)
	return err
}

func (s *Service) LeaveGroup(ctx context.Context, actorID uuid.UUID, groupID int64) error {
	return s.repo.Transaction(ctx, func(repo Repository) error {
		if _, err := repo.FindGroupByID(ctx, groupID); err != nil {
			return err
		}
		member, ok, err := repo.FindMember(ctx, groupID, actorID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNotMember
		}

		if member.Role == RoleOwner {
			count, err := repo.CountMembers(ctx, groupID)
			if err != nil {
				return err
			}
			if count > 1 {
				return ErrOwnerMustTransfer
			}
			if err := repo.RemoveMember(ctx, groupID, actorID); err != nil {
				return err
			}
			_, err = repo.UpdateGroup(ctx, groupID, map[string]any{"status": StatusDissolved})
			return err
		}

		return repo.RemoveMember(ctx, groupID, actorID)
	})
}

func normalizeDescription(description *string) *string {
	if description == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*description)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func generateInviteCode() (string, error) {
	var bytes [9]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:])[:12], nil
}
