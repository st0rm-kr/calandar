package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/google/uuid"
)

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func ptr[T any](value T) *T {
	return &value
}

type fakeScheduleReader struct {
	rows            []schedules.Schedule
	participantRows []schedules.Schedule
	friendRows      []FriendSchedule
	groupRows       []GroupEvent
	friends         []CalendarFriend
	groups          []CalendarGroup
	summaries       map[int64]EventParticipationSummary
}

func (r fakeScheduleReader) ListByRange(_ context.Context, _ uuid.UUID, _ time.Time, _ time.Time) ([]schedules.Schedule, error) {
	return r.rows, nil
}

func (r fakeScheduleReader) ListParticipantEventsByRange(_ context.Context, _ uuid.UUID, _ time.Time, _ time.Time) ([]schedules.Schedule, error) {
	return r.participantRows, nil
}

func (r fakeScheduleReader) ListFriendSchedulesByRange(_ context.Context, _ uuid.UUID, _ time.Time, _ time.Time) ([]FriendSchedule, error) {
	return r.friendRows, nil
}

func (r fakeScheduleReader) ListGroupEventsByRange(_ context.Context, _ uuid.UUID, _ time.Time, _ time.Time) ([]GroupEvent, error) {
	return r.groupRows, nil
}

func (r fakeScheduleReader) ListCalendarFriends(_ context.Context, _ uuid.UUID) ([]CalendarFriend, error) {
	return r.friends, nil
}

func (r fakeScheduleReader) ListCalendarGroups(_ context.Context, _ uuid.UUID) ([]CalendarGroup, error) {
	return r.groups, nil
}

func (r fakeScheduleReader) ListEventParticipationSummaries(_ context.Context, _ uuid.UUID, eventIDs []int64) (map[int64]EventParticipationSummary, error) {
	result := make(map[int64]EventParticipationSummary, len(eventIDs))
	for _, id := range eventIDs {
		if summary, ok := r.summaries[id]; ok {
			result[id] = summary
		}
	}
	return result, nil
}

func TestCalendarSubscriptionsIncludeSelfFriendsAndGroups(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	friendID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	reader := fakeScheduleReader{
		friends: []CalendarFriend{
			{UserID: friendID, DisplayName: "Alex"},
		},
		groups: []CalendarGroup{
			{ID: 7, Name: "攀岩群"},
		},
	}
	service := NewService(reader)

	subscriptions, err := service.Subscriptions(context.Background(), userID)
	if err != nil {
		t.Fatalf("calendar subscriptions: %v", err)
	}

	labels := map[string]string{}
	for _, subscription := range subscriptions {
		labels[subscription.ID] = subscription.Label
	}
	expected := map[string]string{
		"self:schedules": "自身日程",
		"self:events":    "自身活动",
		"friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules": "Alex 日程",
		"group:7:events": "攀岩群 活动",
	}
	for id, label := range expected {
		if labels[id] != label {
			t.Fatalf("expected subscription %s label %q, got labels %+v", id, label, labels)
		}
	}
}

func TestCalendarFiltersBySourceAndMasksFriendBusyOnly(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	friendID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	location := "秘密地点"
	reader := fakeScheduleReader{
		rows: []schedules.Schedule{
			{
				ID:         1,
				UserID:     userID,
				Title:      "Personal run",
				StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
				Visibility: schedules.VisibilityPublic,
				Source:     schedules.SourceManual,
			},
		},
		friendRows: []FriendSchedule{
			{
				Schedule: schedules.Schedule{
					ID:         41,
					UserID:     friendID,
					Title:      "牙医",
					StartAt:    mustTime(t, "2026-06-11T09:00:00Z"),
					EndAt:      ptr(mustTime(t, "2026-06-11T10:00:00Z")),
					Location:   &location,
					Visibility: schedules.VisibilityBusyOnly,
					Source:     schedules.SourceManual,
				},
				FriendID:   friendID,
				FriendName: "Alex",
			},
		},
	}
	service := NewService(reader)

	items, err := service.Month(
		context.Background(),
		userID,
		mustTime(t, "2026-06-01T00:00:00Z"),
		mustTime(t, "2026-07-01T00:00:00Z"),
		"all",
		[]string{"friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules"},
	)
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only selected friend schedule, got %d items: %+v", len(items), items)
	}
	item := items[0]
	if item.Title != "忙碌" {
		t.Fatalf("expected busy_only friend title to be masked, got %q", item.Title)
	}
	if item.Location != nil {
		t.Fatalf("expected busy_only friend location to be hidden, got %q", *item.Location)
	}
	if item.SourceID != "friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules" ||
		item.SourceType != SourceTypeFriendSchedules ||
		item.SourceLabel != "Alex 日程" ||
		item.SourceColor != ColorGray {
		t.Fatalf("expected friend source metadata, got %+v", item)
	}
}

func TestCalendarIncludesManualAndEventSchedules(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(9)
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         1,
			UserID:     userID,
			Title:      "Personal run",
			StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-10T09:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         2,
			UserID:     userID,
			Title:      "Climb event",
			StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-12T09:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceEvent,
			EventID:    &eventID,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 calendar items, got %d", len(items))
	}

	byID := map[int64]CalendarItem{}
	for _, item := range items {
		byID[item.ID] = item
	}
	if byID[1].Kind != "schedule" || byID[1].Color != "green" {
		t.Fatalf("expected manual schedule to be green schedule, got %+v", byID[1])
	}
	if byID[2].Kind != "event" || byID[2].Color != "blue" {
		t.Fatalf("expected event schedule to be blue event, got %+v", byID[2])
	}
}

func TestCalendarIncludesEventParticipationSummary(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	friendID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	eventID := int64(9)
	rsvp := "going"
	reader := fakeScheduleReader{
		rows: []schedules.Schedule{
			{
				ID:         2,
				UserID:     userID,
				Title:      "Climb event",
				StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
				EndAt:      ptr(mustTime(t, "2026-06-12T09:00:00Z")),
				Visibility: schedules.VisibilityPublic,
				Source:     schedules.SourceEvent,
				EventID:    &eventID,
			},
		},
		summaries: map[int64]EventParticipationSummary{
			eventID: {
				EventID:    eventID,
				ViewerRSVP: &rsvp,
				GoingCount: 2,
				ParticipantsPreview: []EventParticipantPreview{
					{ID: userID, DisplayName: "Lily"},
					{ID: friendID, DisplayName: "Alex"},
				},
			},
		},
	}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 calendar item, got %d", len(items))
	}

	item := items[0]
	if item.ViewerRSVP == nil || *item.ViewerRSVP != "going" {
		t.Fatalf("expected viewer rsvp going, got %+v", item.ViewerRSVP)
	}
	if item.GoingCount != 2 {
		t.Fatalf("expected going count 2, got %d", item.GoingCount)
	}
	if len(item.ParticipantsPreview) != 2 || item.ParticipantsPreview[0].DisplayName != "Lily" {
		t.Fatalf("expected participant preview to be included, got %+v", item.ParticipantsPreview)
	}
}

func TestCalendarIncludesDismissedParticipantEvents(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(9)
	rsvp := "not_going"
	reader := fakeScheduleReader{
		participantRows: []schedules.Schedule{
			{
				ID:         88,
				UserID:     userID,
				Title:      "Climb event",
				StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
				EndAt:      ptr(mustTime(t, "2026-06-12T09:00:00Z")),
				Visibility: schedules.VisibilityPublic,
				Source:     schedules.SourceEvent,
				EventID:    &eventID,
			},
		},
		summaries: map[int64]EventParticipationSummary{
			eventID: {
				EventID:    eventID,
				ViewerRSVP: &rsvp,
				GoingCount: 2,
			},
		},
	}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected dismissed event to stay in calendar feed, got %d items", len(items))
	}
	if items[0].EventID == nil || *items[0].EventID != eventID {
		t.Fatalf("expected dismissed event item, got %+v", items[0])
	}
	if items[0].ViewerRSVP == nil || *items[0].ViewerRSVP != "not_going" {
		t.Fatalf("expected viewer rsvp not_going, got %+v", items[0].ViewerRSVP)
	}
}

func TestCalendarMarksBusyOnlyAsGray(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         3,
			UserID:     userID,
			Title:      "Busy block",
			StartAt:    mustTime(t, "2026-06-15T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-15T09:00:00Z")),
			Visibility: schedules.VisibilityBusyOnly,
			Source:     schedules.SourceManual,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Color != "gray" {
		t.Fatalf("expected busy_only item to be gray, got %q", items[0].Color)
	}
}

func TestCalendarMarksConflictingItems(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         4,
			UserID:     userID,
			Title:      "A",
			StartAt:    mustTime(t, "2026-06-20T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-20T10:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         5,
			UserID:     userID,
			Title:      "B",
			StartAt:    mustTime(t, "2026-06-20T09:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-20T11:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	for _, item := range items {
		if !item.HasConflict {
			t.Fatalf("expected overlapping item %d to be marked as conflict", item.ID)
		}
	}
}

func TestCalendarAcceptsGroupsFilter(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(9)
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         6,
			UserID:     userID,
			Title:      "Personal",
			StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         7,
			UserID:     userID,
			Title:      "Group event",
			StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceEvent,
			EventID:    &eventID,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "groups")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only event items with groups filter, got %d", len(items))
	}
	if items[0].ID != 7 {
		t.Fatalf("expected event item, got %+v", items[0])
	}
}
