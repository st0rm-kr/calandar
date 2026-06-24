package events

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type fakeFriendAuthorizer struct {
	friends map[string]bool
}

func (f *fakeFriendAuthorizer) AreFriends(_ context.Context, a, b uuid.UUID) (bool, error) {
	return f.friends[friendPairKey(a, b)], nil
}

func friendPairKey(a, b uuid.UUID) string {
	if a.String() < b.String() {
		return a.String() + ":" + b.String()
	}
	return b.String() + ":" + a.String()
}

type fakeGroupAuthorizer struct {
	members map[string]bool
}

func (g *fakeGroupAuthorizer) IsMember(_ context.Context, groupID int64, userID uuid.UUID) (bool, error) {
	return g.members[fmt.Sprintf("%d:%s", groupID, userID.String())], nil
}

func TestInvitePersonalEventInvitesAcceptedFriends(t *testing.T) {
	repo := newFakeEventRepository()
	owner := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	friend := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	event := activeEvent()
	event.OwnerID = owner
	event.Scope = "personal"
	event = repo.seedEvent(event)

	service := NewService(repo, WithFriendAuthorizer(&fakeFriendAuthorizer{
		friends: map[string]bool{friendPairKey(owner, friend): true},
	}))

	invited, err := service.Invite(context.Background(), owner, event.ID, []uuid.UUID{friend})
	if err != nil {
		t.Fatalf("invite friend: %v", err)
	}
	if len(invited) != 1 {
		t.Fatalf("expected 1 invited participant, got %d", len(invited))
	}
	participant := repo.participants[participantKey(event.ID, friend)]
	if participant.RSVP != RSVPInvited || participant.Source != ParticipantSourceInvited {
		t.Fatalf("expected invited/invited participant, got %+v", participant)
	}
}

func TestInviteGroupEventInvitesGroupMembers(t *testing.T) {
	repo := newFakeEventRepository()
	creator := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	member := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	groupID := int64(7)
	event := activeEvent()
	event.OwnerID = creator
	event.Scope = "group"
	event.GroupID = &groupID
	event = repo.seedEvent(event)

	service := NewService(repo, WithGroupAuthorizer(&fakeGroupAuthorizer{
		members: map[string]bool{
			fmt.Sprintf("%d:%s", groupID, creator.String()): true,
			fmt.Sprintf("%d:%s", groupID, member.String()):  true,
		},
	}))

	invited, err := service.Invite(context.Background(), creator, event.ID, []uuid.UUID{member})
	if err != nil {
		t.Fatalf("invite group member: %v", err)
	}
	if len(invited) != 1 {
		t.Fatalf("expected 1 invited participant, got %d", len(invited))
	}
	participant := repo.participants[participantKey(event.ID, member)]
	if participant.RSVP != RSVPInvited || participant.Source != ParticipantSourceInvited {
		t.Fatalf("expected invited/invited participant, got %+v", participant)
	}
}

func TestInviteKeepsExistingParticipantUnchanged(t *testing.T) {
	repo := newFakeEventRepository()
	owner := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	friend := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	event := activeEvent()
	event.OwnerID = owner
	event.Scope = "personal"
	event = repo.seedEvent(event)
	repo.participants[participantKey(event.ID, friend)] = Participant{
		ID:      55,
		EventID: event.ID,
		UserID:  friend,
		RSVP:    RSVPGoing,
		Source:  ParticipantSourceSelf,
	}

	service := NewService(repo, WithFriendAuthorizer(&fakeFriendAuthorizer{
		friends: map[string]bool{friendPairKey(owner, friend): true},
	}))

	if _, err := service.Invite(context.Background(), owner, event.ID, []uuid.UUID{friend}); err != nil {
		t.Fatalf("invite existing participant: %v", err)
	}

	participant := repo.participants[participantKey(event.ID, friend)]
	if participant.ID != 55 || participant.RSVP != RSVPGoing || participant.Source != ParticipantSourceSelf {
		t.Fatalf("expected existing going participant unchanged, got %+v", participant)
	}
}

func TestInviteNonFriendPersonalEventReturnsForbidden(t *testing.T) {
	repo := newFakeEventRepository()
	owner := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	stranger := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	event := activeEvent()
	event.OwnerID = owner
	event.Scope = "personal"
	event = repo.seedEvent(event)

	service := NewService(repo, WithFriendAuthorizer(&fakeFriendAuthorizer{friends: map[string]bool{}}))

	_, err := service.Invite(context.Background(), owner, event.ID, []uuid.UUID{stranger})
	if !errors.Is(err, ErrInviteForbidden) {
		t.Fatalf("expected ErrInviteForbidden, got %v", err)
	}
	if len(repo.participants) != 0 {
		t.Fatalf("expected no participant inserted, got %d", len(repo.participants))
	}
}

func TestInviteNonOwnerPersonalEventReturnsForbidden(t *testing.T) {
	repo := newFakeEventRepository()
	owner := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	other := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	friend := uuid.MustParse("cccccccc-cccc-4ccc-cccc-cccccccccccc")
	event := activeEvent()
	event.OwnerID = owner
	event.Scope = "personal"
	event = repo.seedEvent(event)

	service := NewService(repo, WithFriendAuthorizer(&fakeFriendAuthorizer{
		friends: map[string]bool{friendPairKey(other, friend): true},
	}))

	_, err := service.Invite(context.Background(), other, event.ID, []uuid.UUID{friend})
	if !errors.Is(err, ErrInviteForbidden) {
		t.Fatalf("expected ErrInviteForbidden for non-owner, got %v", err)
	}
}

func TestInviteGroupEventByNonMemberReturnsForbidden(t *testing.T) {
	repo := newFakeEventRepository()
	creator := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	outsider := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	member := uuid.MustParse("cccccccc-cccc-4ccc-cccc-cccccccccccc")
	groupID := int64(7)
	event := activeEvent()
	event.OwnerID = creator
	event.Scope = "group"
	event.GroupID = &groupID
	event = repo.seedEvent(event)

	service := NewService(repo, WithGroupAuthorizer(&fakeGroupAuthorizer{
		members: map[string]bool{fmt.Sprintf("%d:%s", groupID, member.String()): true},
	}))

	_, err := service.Invite(context.Background(), outsider, event.ID, []uuid.UUID{member})
	if !errors.Is(err, ErrInviteForbidden) {
		t.Fatalf("expected ErrInviteForbidden for non-member actor, got %v", err)
	}
}
