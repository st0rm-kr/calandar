package groups

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

var (
	owner   = uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	member1 = uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	member2 = uuid.MustParse("cccccccc-cccc-4ccc-cccc-cccccccccccc")
)

func TestCreateGroupInsertsOwnerMember(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, err := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if group.OwnerID != owner {
		t.Fatalf("expected owner %v, got %v", owner, group.OwnerID)
	}
	if group.InviteCode == "" {
		t.Fatalf("expected invite code to be generated")
	}

	mem, found, err := repo.FindMember(context.Background(), group.ID, owner)
	if err != nil {
		t.Fatalf("find member: %v", err)
	}
	if !found || mem.Role != RoleOwner {
		t.Fatalf("expected owner member row, found=%v role=%q", found, mem.Role)
	}
}

func TestJoinByInviteCodeInsertsMemberOnce(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})

	if _, err := service.JoinByInviteCode(context.Background(), member1, group.InviteCode); err != nil {
		t.Fatalf("first join: %v", err)
	}
	if _, err := service.JoinByInviteCode(context.Background(), member1, group.InviteCode); err != nil {
		t.Fatalf("second join: %v", err)
	}

	if got := repo.memberCount(group.ID); got != 2 {
		t.Fatalf("expected 2 members (owner+member1), got %d", got)
	}
}

func TestOnlyOwnerCanPatchAndDissolve(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	if _, err := service.JoinByInviteCode(context.Background(), member1, group.InviteCode); err != nil {
		t.Fatalf("join: %v", err)
	}

	newName := "Boulderers"
	if _, err := service.UpdateGroup(context.Background(), member1, group.ID, UpdateGroupInput{Name: &newName}); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner on patch by member, got %v", err)
	}
	if err := service.DissolveGroup(context.Background(), member1, group.ID); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner on dissolve by member, got %v", err)
	}

	updated, err := service.UpdateGroup(context.Background(), owner, group.ID, UpdateGroupInput{Name: &newName})
	if err != nil {
		t.Fatalf("owner patch: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("expected name %q, got %q", newName, updated.Name)
	}
}

func TestInviteCreatesPendingInvite(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})

	invite, err := service.Invite(context.Background(), owner, group.ID, member1)
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if invite.Status != InviteStatusPending {
		t.Fatalf("expected pending invite, got %q", invite.Status)
	}
	if invite.GroupID != group.ID || invite.InviteeID != member1 {
		t.Fatalf("unexpected invite %+v", invite)
	}
}

func TestNonMemberCannotInvite(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})

	if _, err := service.Invite(context.Background(), member1, group.ID, member2); !errors.Is(err, ErrNotMember) {
		t.Fatalf("expected ErrNotMember when non-member invites, got %v", err)
	}
}

func TestAcceptInviteInsertsMemberAndMarksAccepted(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	invite, _ := service.Invite(context.Background(), owner, group.ID, member1)

	if err := service.AcceptInvite(context.Background(), member1, invite.ID); err != nil {
		t.Fatalf("accept invite: %v", err)
	}

	_, found, err := repo.FindMember(context.Background(), group.ID, member1)
	if err != nil {
		t.Fatalf("find member: %v", err)
	}
	if !found {
		t.Fatalf("expected member1 to be a member after accept")
	}
	stored := repo.invites[invite.ID]
	if stored.Status != InviteStatusAccepted {
		t.Fatalf("expected invite accepted, got %q", stored.Status)
	}
}

func TestAcceptInviteRejectsNonInvitee(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	invite, _ := service.Invite(context.Background(), owner, group.ID, member1)

	if err := service.AcceptInvite(context.Background(), member2, invite.ID); !errors.Is(err, ErrNotAuthorized) {
		t.Fatalf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestOwnerLeavingLastGroupDissolves(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})

	if err := service.LeaveGroup(context.Background(), owner, group.ID); err != nil {
		t.Fatalf("owner leave last: %v", err)
	}

	stored := repo.groups[group.ID]
	if stored.Status != StatusDissolved {
		t.Fatalf("expected group dissolved, got %q", stored.Status)
	}
	if repo.memberCount(group.ID) != 0 {
		t.Fatalf("expected no members after dissolve, got %d", repo.memberCount(group.ID))
	}
}

func TestOwnerLeavingNonEmptyGroupRequiresTransfer(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	if _, err := service.JoinByInviteCode(context.Background(), member1, group.InviteCode); err != nil {
		t.Fatalf("join: %v", err)
	}

	if err := service.LeaveGroup(context.Background(), owner, group.ID); !errors.Is(err, ErrOwnerMustTransfer) {
		t.Fatalf("expected ErrOwnerMustTransfer, got %v", err)
	}

	stored := repo.groups[group.ID]
	if stored.Status != StatusActive {
		t.Fatalf("expected group to stay active, got %q", stored.Status)
	}
}

func TestMemberCanLeaveGroup(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	group, _ := service.CreateGroup(context.Background(), owner, CreateGroupInput{Name: "Climbers"})
	if _, err := service.JoinByInviteCode(context.Background(), member1, group.InviteCode); err != nil {
		t.Fatalf("join: %v", err)
	}

	if err := service.LeaveGroup(context.Background(), member1, group.ID); err != nil {
		t.Fatalf("member leave: %v", err)
	}
	if repo.memberCount(group.ID) != 1 {
		t.Fatalf("expected 1 member after leave, got %d", repo.memberCount(group.ID))
	}
	stored := repo.groups[group.ID]
	if stored.Status != StatusActive {
		t.Fatalf("expected group active after member leaves, got %q", stored.Status)
	}
}

type fakeRepo struct {
	groups       map[int64]Group
	members      map[int64]GroupMember
	invites      map[int64]GroupInvite
	nextGroupID  int64
	nextMemberID int64
	nextInviteID int64
	profiles     map[uuid.UUID]Member
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		groups:  map[int64]Group{},
		members: map[int64]GroupMember{},
		invites: map[int64]GroupInvite{},
		profiles: map[uuid.UUID]Member{
			owner:   {UserID: owner, DisplayName: "Owner"},
			member1: {UserID: member1, DisplayName: "Member1"},
			member2: {UserID: member2, DisplayName: "Member2"},
		},
	}
}

func (r *fakeRepo) memberCount(groupID int64) int {
	count := 0
	for _, m := range r.members {
		if m.GroupID == groupID && m.DeletedAt == nil {
			count++
		}
	}
	return count
}

func (r *fakeRepo) Transaction(ctx context.Context, fn func(Repository) error) error {
	return fn(r)
}

func (r *fakeRepo) CreateGroup(_ context.Context, group Group) (Group, error) {
	r.nextGroupID++
	group.ID = r.nextGroupID
	r.groups[group.ID] = group
	return group, nil
}

func (r *fakeRepo) FindGroupByID(_ context.Context, id int64) (Group, error) {
	group, ok := r.groups[id]
	if !ok || group.DeletedAt != nil {
		return Group{}, ErrGroupNotFound
	}
	return group, nil
}

func (r *fakeRepo) FindGroupByInviteCode(_ context.Context, code string) (Group, error) {
	for _, g := range r.groups {
		if g.InviteCode == code && g.DeletedAt == nil {
			return g, nil
		}
	}
	return Group{}, ErrGroupNotFound
}

func (r *fakeRepo) UpdateGroup(_ context.Context, id int64, fields map[string]any) (Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return Group{}, ErrGroupNotFound
	}
	if name, ok := fields["name"].(string); ok {
		group.Name = name
	}
	if desc, ok := fields["description"]; ok {
		if desc == nil {
			group.Description = nil
		} else if s, ok := desc.(string); ok {
			group.Description = &s
		}
	}
	if status, ok := fields["status"].(string); ok {
		group.Status = status
	}
	r.groups[id] = group
	return group, nil
}

func (r *fakeRepo) ListGroupsByMember(_ context.Context, userID uuid.UUID) ([]Group, error) {
	var result []Group
	for _, m := range r.members {
		if m.UserID == userID && m.DeletedAt == nil {
			if g, ok := r.groups[m.GroupID]; ok && g.DeletedAt == nil {
				result = append(result, g)
			}
		}
	}
	return result, nil
}

func (r *fakeRepo) FindMember(_ context.Context, groupID int64, userID uuid.UUID) (GroupMember, bool, error) {
	for _, m := range r.members {
		if m.GroupID == groupID && m.UserID == userID && m.DeletedAt == nil {
			return m, true, nil
		}
	}
	return GroupMember{}, false, nil
}

func (r *fakeRepo) AddMember(_ context.Context, member GroupMember) (GroupMember, error) {
	r.nextMemberID++
	member.ID = r.nextMemberID
	r.members[member.ID] = member
	return member, nil
}

func (r *fakeRepo) RemoveMember(_ context.Context, groupID int64, userID uuid.UUID) error {
	for id, m := range r.members {
		if m.GroupID == groupID && m.UserID == userID && m.DeletedAt == nil {
			delete(r.members, id)
		}
	}
	return nil
}

func (r *fakeRepo) CountMembers(_ context.Context, groupID int64) (int, error) {
	return r.memberCount(groupID), nil
}

func (r *fakeRepo) ListMembers(_ context.Context, groupID int64) ([]Member, error) {
	var result []Member
	for _, m := range r.members {
		if m.GroupID == groupID && m.DeletedAt == nil {
			profile := r.profiles[m.UserID]
			profile.UserID = m.UserID
			profile.Role = m.Role
			result = append(result, profile)
		}
	}
	return result, nil
}

func (r *fakeRepo) CreateInvite(_ context.Context, invite GroupInvite) (GroupInvite, error) {
	r.nextInviteID++
	invite.ID = r.nextInviteID
	r.invites[invite.ID] = invite
	return invite, nil
}

func (r *fakeRepo) FindInviteByID(_ context.Context, id int64) (GroupInvite, error) {
	invite, ok := r.invites[id]
	if !ok || invite.DeletedAt != nil {
		return GroupInvite{}, ErrInviteNotFound
	}
	return invite, nil
}

func (r *fakeRepo) FindInvite(_ context.Context, groupID int64, inviteeID uuid.UUID) (GroupInvite, bool, error) {
	for _, inv := range r.invites {
		if inv.GroupID == groupID && inv.InviteeID == inviteeID && inv.DeletedAt == nil {
			return inv, true, nil
		}
	}
	return GroupInvite{}, false, nil
}

func (r *fakeRepo) UpdateInviteStatus(_ context.Context, id int64, status string) (GroupInvite, error) {
	invite, ok := r.invites[id]
	if !ok {
		return GroupInvite{}, ErrInviteNotFound
	}
	invite.Status = status
	r.invites[id] = invite
	return invite, nil
}
