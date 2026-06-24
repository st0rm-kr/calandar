package friends

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

var (
	userA = uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	userB = uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	userC = uuid.MustParse("cccccccc-cccc-4ccc-cccc-cccccccccccc")
)

func TestRequestRejectsSelf(t *testing.T) {
	service := NewService(newFakeRepo())
	_, err := service.SendRequest(context.Background(), userA, userA)
	if !errors.Is(err, ErrCannotFriendSelf) {
		t.Fatalf("expected ErrCannotFriendSelf, got %v", err)
	}
}

func TestRequestIsIdempotentWhilePending(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	first, err := service.SendRequest(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	second, err := service.SendRequest(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected idempotent request, got ids %d and %d", first.ID, second.ID)
	}
	if repo.count() != 1 {
		t.Fatalf("expected one friendship row, got %d", repo.count())
	}
}

func TestReversePendingRequestAcceptCreatesAcceptedRelationship(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	request, err := service.SendRequest(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}

	accepted, err := service.Accept(context.Background(), userB, request.ID)
	if err != nil {
		t.Fatalf("accept request: %v", err)
	}
	if accepted.Status != StatusAccepted {
		t.Fatalf("expected accepted status, got %q", accepted.Status)
	}
}

func TestAcceptedFriendshipAppearsInBothLists(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	request, _ := service.SendRequest(context.Background(), userA, userB)
	if _, err := service.Accept(context.Background(), userB, request.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}

	aList, err := service.ListFriends(context.Background(), userA)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	bList, err := service.ListFriends(context.Background(), userB)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(aList) != 1 || aList[0].UserID != userB {
		t.Fatalf("expected B in A's list, got %+v", aList)
	}
	if len(bList) != 1 || bList[0].UserID != userA {
		t.Fatalf("expected A in B's list, got %+v", bList)
	}
}

func TestDeleteFriendRemovesRelationshipForBoth(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	request, _ := service.SendRequest(context.Background(), userA, userB)
	if _, err := service.Accept(context.Background(), userB, request.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}

	if err := service.DeleteFriend(context.Background(), userA, userB); err != nil {
		t.Fatalf("delete friend: %v", err)
	}

	aList, _ := service.ListFriends(context.Background(), userA)
	bList, _ := service.ListFriends(context.Background(), userB)
	if len(aList) != 0 || len(bList) != 0 {
		t.Fatalf("expected empty lists after delete, got %d and %d", len(aList), len(bList))
	}
}

func TestRejectedRequestCanBeReSentByRequester(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	request, _ := service.SendRequest(context.Background(), userA, userB)
	if _, err := service.Reject(context.Background(), userB, request.ID); err != nil {
		t.Fatalf("reject: %v", err)
	}

	again, err := service.SendRequest(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("re-send after reject: %v", err)
	}
	if again.Status != StatusPending {
		t.Fatalf("expected pending after re-send, got %q", again.Status)
	}
}

func TestAcceptedCannotBeRequestedAgain(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	request, _ := service.SendRequest(context.Background(), userA, userB)
	if _, err := service.Accept(context.Background(), userB, request.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}

	_, err := service.SendRequest(context.Background(), userA, userB)
	if !errors.Is(err, ErrAlreadyFriends) {
		t.Fatalf("expected ErrAlreadyFriends, got %v", err)
	}
}

type fakeFriendsRepo struct {
	rows    map[int64]Friendship
	nextID  int64
	profile map[uuid.UUID]Friend
}

func newFakeRepo() *fakeFriendsRepo {
	return &fakeFriendsRepo{
		rows:   map[int64]Friendship{},
		nextID: 0,
		profile: map[uuid.UUID]Friend{
			userA: {UserID: userA, DisplayName: "A"},
			userB: {UserID: userB, DisplayName: "B"},
			userC: {UserID: userC, DisplayName: "C"},
		},
	}
}

func (r *fakeFriendsRepo) count() int {
	return len(r.rows)
}

func (r *fakeFriendsRepo) FindBetween(_ context.Context, a, b uuid.UUID) (Friendship, bool, error) {
	for _, row := range r.rows {
		if (row.RequesterID == a && row.AddresseeID == b) || (row.RequesterID == b && row.AddresseeID == a) {
			return row, true, nil
		}
	}
	return Friendship{}, false, nil
}

func (r *fakeFriendsRepo) Create(_ context.Context, friendship Friendship) (Friendship, error) {
	r.nextID++
	friendship.ID = r.nextID
	r.rows[friendship.ID] = friendship
	return friendship, nil
}

func (r *fakeFriendsRepo) UpdateStatus(_ context.Context, id int64, status string) (Friendship, error) {
	row, ok := r.rows[id]
	if !ok {
		return Friendship{}, ErrRequestNotFound
	}
	row.Status = status
	r.rows[id] = row
	return row, nil
}

func (r *fakeFriendsRepo) FindByID(_ context.Context, id int64) (Friendship, error) {
	row, ok := r.rows[id]
	if !ok {
		return Friendship{}, ErrRequestNotFound
	}
	return row, nil
}

func (r *fakeFriendsRepo) ListAccepted(_ context.Context, userID uuid.UUID) ([]Friend, error) {
	var friends []Friend
	for _, row := range r.rows {
		if row.Status != StatusAccepted {
			continue
		}
		switch userID {
		case row.RequesterID:
			friends = append(friends, r.profile[row.AddresseeID])
		case row.AddresseeID:
			friends = append(friends, r.profile[row.RequesterID])
		}
	}
	return friends, nil
}

func (r *fakeFriendsRepo) ListRequests(_ context.Context, userID uuid.UUID) ([]Friendship, error) {
	var requests []Friendship
	for _, row := range r.rows {
		if row.Status == StatusPending && (row.RequesterID == userID || row.AddresseeID == userID) {
			requests = append(requests, row)
		}
	}
	return requests, nil
}

func (r *fakeFriendsRepo) DeleteBetween(_ context.Context, a, b uuid.UUID) error {
	for id, row := range r.rows {
		if (row.RequesterID == a && row.AddresseeID == b) || (row.RequesterID == b && row.AddresseeID == a) {
			delete(r.rows, id)
		}
	}
	return nil
}
