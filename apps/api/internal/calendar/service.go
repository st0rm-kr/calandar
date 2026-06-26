package calendar

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/google/uuid"
)

const (
	FilterAll     = "all"
	FilterGroups  = "groups"
	FilterFriends = "friends"

	KindSchedule = "schedule"
	KindEvent    = "event"

	ColorBlue  = "blue"
	ColorGreen = "green"
	ColorGray  = "gray"

	SourceIDSelfSchedules = "self:schedules"
	SourceIDSelfEvents    = "self:events"

	SourceTypeSelfSchedules   = "self_schedules"
	SourceTypeSelfEvents      = "self_events"
	SourceTypeFriendSchedules = "friend_schedules"
	SourceTypeGroupEvents     = "group_events"
)

type ScheduleReader interface {
	ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error)
}

type ParticipationReader interface {
	ListParticipantEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error)
	ListEventParticipationSummaries(ctx context.Context, userID uuid.UUID, eventIDs []int64) (map[int64]schedules.EventParticipationSummary, error)
}

type SourceReader interface {
	ListFriendSchedulesByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]FriendSchedule, error)
	ListGroupEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]GroupEvent, error)
	ListCalendarFriends(ctx context.Context, userID uuid.UUID) ([]CalendarFriend, error)
	ListCalendarGroups(ctx context.Context, userID uuid.UUID) ([]CalendarGroup, error)
}

type CalendarReader interface {
	ScheduleReader
	ParticipationReader
	SourceReader
}

type EventParticipantPreview = schedules.EventParticipantPreview
type EventParticipationSummary = schedules.EventParticipationSummary

type CalendarSubscription struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Label   string `json:"label"`
	Color   string `json:"color"`
	Enabled bool   `json:"enabled"`
}

type CalendarFriend struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

type CalendarGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type FriendSchedule struct {
	Schedule   schedules.Schedule
	FriendID   uuid.UUID
	FriendName string
}

type GroupEvent struct {
	Schedule  schedules.Schedule
	GroupID   int64
	GroupName string
}

type calendarRow struct {
	schedule    schedules.Schedule
	sourceID    string
	sourceType  string
	sourceLabel string
	sourceColor string
}

type CalendarItem struct {
	ID                  int64                     `json:"id"`
	Kind                string                    `json:"kind"`
	Title               string                    `json:"title"`
	StartAt             time.Time                 `json:"start_at"`
	EndAt               *time.Time                `json:"end_at"`
	Location            *string                   `json:"location"`
	Visibility          string                    `json:"visibility"`
	Color               string                    `json:"color"`
	HasConflict         bool                      `json:"has_conflict"`
	EventID             *int64                    `json:"event_id"`
	ViewerRSVP          *string                   `json:"viewer_rsvp"`
	GoingCount          int                       `json:"going_count"`
	SourceID            string                    `json:"source_id"`
	SourceType          string                    `json:"source_type"`
	SourceLabel         string                    `json:"source_label"`
	SourceColor         string                    `json:"source_color"`
	ParticipantsPreview []EventParticipantPreview `json:"participants_preview"`
}

type Service struct {
	reader CalendarReader
}

func NewService(reader CalendarReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) Subscriptions(ctx context.Context, userID uuid.UUID) ([]CalendarSubscription, error) {
	friends, err := s.reader.ListCalendarFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	groups, err := s.reader.ListCalendarGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(friends, func(i, j int) bool {
		if friends[i].DisplayName == friends[j].DisplayName {
			return friends[i].UserID.String() < friends[j].UserID.String()
		}
		return friends[i].DisplayName < friends[j].DisplayName
	})
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Name == groups[j].Name {
			return groups[i].ID < groups[j].ID
		}
		return groups[i].Name < groups[j].Name
	})

	subscriptions := []CalendarSubscription{
		{
			ID:      SourceIDSelfSchedules,
			Type:    SourceTypeSelfSchedules,
			Label:   "自身日程",
			Color:   ColorGreen,
			Enabled: true,
		},
		{
			ID:      SourceIDSelfEvents,
			Type:    SourceTypeSelfEvents,
			Label:   "自身活动",
			Color:   ColorBlue,
			Enabled: true,
		},
	}
	for _, friend := range friends {
		subscriptions = append(subscriptions, CalendarSubscription{
			ID:      friendScheduleSourceID(friend.UserID),
			Type:    SourceTypeFriendSchedules,
			Label:   sourceLabel(friend.DisplayName, "好友", " 日程"),
			Color:   ColorGray,
			Enabled: true,
		})
	}
	for _, group := range groups {
		subscriptions = append(subscriptions, CalendarSubscription{
			ID:      groupEventSourceID(group.ID),
			Type:    SourceTypeGroupEvents,
			Label:   sourceLabel(group.Name, "群组", " 活动"),
			Color:   ColorBlue,
			Enabled: true,
		})
	}
	return subscriptions, nil
}

func (s *Service) Month(ctx context.Context, userID uuid.UUID, from, to time.Time, filter string, sources ...[]string) ([]CalendarItem, error) {
	var selectedSourceIDs []string
	if len(sources) > 0 {
		selectedSourceIDs = sources[0]
	}
	selectedSources := selectedSourceSet(selectedSourceIDs)

	rows, err := s.reader.ListByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	participantRows, err := s.reader.ListParticipantEventsByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	friendRows, err := s.reader.ListFriendSchedulesByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	groupRows, err := s.reader.ListGroupEventsByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}

	calendarRows := make([]calendarRow, 0, len(rows)+len(participantRows)+len(friendRows)+len(groupRows))
	for _, row := range rows {
		calendarRows = append(calendarRows, selfCalendarRow(row))
	}
	for _, row := range participantRows {
		calendarRows = append(calendarRows, selfEventCalendarRow(row))
	}
	for _, row := range friendRows {
		calendarRows = append(calendarRows, friendCalendarRow(row))
	}
	for _, row := range groupRows {
		calendarRows = append(calendarRows, groupCalendarRow(row))
	}
	calendarRows = filterCalendarRows(mergeCalendarRows(calendarRows), filter, selectedSources)

	eventIDs := eventIDsFromRows(calendarRows)
	summaries, err := s.reader.ListEventParticipationSummaries(ctx, userID, eventIDs)
	if err != nil {
		return nil, err
	}

	items := make([]CalendarItem, 0, len(calendarRows))
	for _, row := range calendarRows {
		schedule := row.schedule
		if row.sourceType == SourceTypeFriendSchedules && schedule.Visibility == schedules.VisibilityPrivate {
			continue
		}
		kind := KindSchedule
		if schedule.Source == schedules.SourceEvent {
			kind = KindEvent
		}
		title := schedule.Title
		location := schedule.Location
		if row.sourceType == SourceTypeFriendSchedules && schedule.Visibility == schedules.VisibilityBusyOnly {
			title = "忙碌"
			location = nil
		}
		item := CalendarItem{
			ID:                  schedule.ID,
			Kind:                kind,
			Title:               title,
			StartAt:             schedule.StartAt,
			EndAt:               schedule.EndAt,
			Location:            location,
			Visibility:          schedule.Visibility,
			Color:               itemColor(kind, schedule.Visibility),
			EventID:             schedule.EventID,
			SourceID:            row.sourceID,
			SourceType:          row.sourceType,
			SourceLabel:         row.sourceLabel,
			SourceColor:         row.sourceColor,
			ParticipantsPreview: []EventParticipantPreview{},
		}
		if schedule.EventID != nil {
			if summary, ok := summaries[*schedule.EventID]; ok {
				item.ViewerRSVP = summary.ViewerRSVP
				item.GoingCount = summary.GoingCount
				item.ParticipantsPreview = summary.ParticipantsPreview
				if item.ParticipantsPreview == nil {
					item.ParticipantsPreview = []EventParticipantPreview{}
				}
			}
		}
		items = append(items, item)
	}

	markConflicts(items)
	return items, nil
}

func mergeCalendarRows(input []calendarRow) []calendarRow {
	rows := make([]calendarRow, 0, len(input))
	seenEventIDs := map[int64]bool{}
	for _, row := range input {
		if row.schedule.Source == schedules.SourceEvent && row.schedule.EventID != nil {
			if seenEventIDs[*row.schedule.EventID] {
				continue
			}
			seenEventIDs[*row.schedule.EventID] = true
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].schedule.StartAt.Before(rows[j].schedule.StartAt)
	})
	return rows
}

func eventIDsFromRows(rows []calendarRow) []int64 {
	seen := map[int64]bool{}
	eventIDs := make([]int64, 0)
	for _, row := range rows {
		if row.schedule.Source != schedules.SourceEvent || row.schedule.EventID == nil {
			continue
		}
		if seen[*row.schedule.EventID] {
			continue
		}
		seen[*row.schedule.EventID] = true
		eventIDs = append(eventIDs, *row.schedule.EventID)
	}
	return eventIDs
}

func filterCalendarRows(rows []calendarRow, filter string, selectedSources map[string]bool) []calendarRow {
	filtered := rows[:0]
	for _, row := range rows {
		kind := KindSchedule
		if row.schedule.Source == schedules.SourceEvent {
			kind = KindEvent
		}
		if !includeForFilter(filter, kind, row.sourceType) {
			continue
		}
		if selectedSources != nil && !selectedSources[row.sourceID] {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func includeForFilter(filter, kind, sourceType string) bool {
	switch filter {
	case FilterGroups:
		return kind == KindEvent
	case FilterFriends:
		return sourceType == SourceTypeFriendSchedules
	default:
		return true
	}
}

func selectedSourceSet(sourceIDs []string) map[string]bool {
	if sourceIDs == nil {
		return nil
	}
	selected := make(map[string]bool, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		if sourceID == "" {
			continue
		}
		selected[sourceID] = true
	}
	return selected
}

func selfCalendarRow(schedule schedules.Schedule) calendarRow {
	if schedule.Source == schedules.SourceEvent {
		return selfEventCalendarRow(schedule)
	}
	return calendarRow{
		schedule:    schedule,
		sourceID:    SourceIDSelfSchedules,
		sourceType:  SourceTypeSelfSchedules,
		sourceLabel: "自身日程",
		sourceColor: ColorGreen,
	}
}

func selfEventCalendarRow(schedule schedules.Schedule) calendarRow {
	return calendarRow{
		schedule:    schedule,
		sourceID:    SourceIDSelfEvents,
		sourceType:  SourceTypeSelfEvents,
		sourceLabel: "自身活动",
		sourceColor: ColorBlue,
	}
}

func friendCalendarRow(row FriendSchedule) calendarRow {
	return calendarRow{
		schedule:    row.Schedule,
		sourceID:    friendScheduleSourceID(row.FriendID),
		sourceType:  SourceTypeFriendSchedules,
		sourceLabel: sourceLabel(row.FriendName, "好友", " 日程"),
		sourceColor: ColorGray,
	}
}

func groupCalendarRow(row GroupEvent) calendarRow {
	return calendarRow{
		schedule:    row.Schedule,
		sourceID:    groupEventSourceID(row.GroupID),
		sourceType:  SourceTypeGroupEvents,
		sourceLabel: sourceLabel(row.GroupName, "群组", " 活动"),
		sourceColor: ColorBlue,
	}
}

func friendScheduleSourceID(friendID uuid.UUID) string {
	return "friend:" + friendID.String() + ":schedules"
}

func groupEventSourceID(groupID int64) string {
	return "group:" + strconv.FormatInt(groupID, 10) + ":events"
}

func sourceLabel(name, fallback, suffix string) string {
	if name == "" {
		return fallback + suffix
	}
	return name + suffix
}

func itemColor(kind, visibility string) string {
	if visibility == schedules.VisibilityBusyOnly {
		return ColorGray
	}
	if kind == KindEvent {
		return ColorBlue
	}
	return ColorGreen
}

func markConflicts(items []CalendarItem) {
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if overlaps(items[i], items[j]) {
				items[i].HasConflict = true
				items[j].HasConflict = true
			}
		}
	}
}

func overlaps(a, b CalendarItem) bool {
	return a.StartAt.Before(effectiveEnd(b)) && b.StartAt.Before(effectiveEnd(a))
}

func effectiveEnd(item CalendarItem) time.Time {
	if item.EndAt != nil {
		return *item.EndAt
	}
	return item.StartAt.Add(time.Minute)
}
