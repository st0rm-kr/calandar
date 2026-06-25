package calendar

import (
	"context"
	"sort"
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
)

type ScheduleReader interface {
	ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error)
}

type ParticipationReader interface {
	ListParticipantEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error)
	ListEventParticipationSummaries(ctx context.Context, userID uuid.UUID, eventIDs []int64) (map[int64]schedules.EventParticipationSummary, error)
}

type CalendarReader interface {
	ScheduleReader
	ParticipationReader
}

type EventParticipantPreview = schedules.EventParticipantPreview
type EventParticipationSummary = schedules.EventParticipationSummary

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
	ParticipantsPreview []EventParticipantPreview `json:"participants_preview"`
}

type Service struct {
	reader CalendarReader
}

func NewService(reader CalendarReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) Month(ctx context.Context, userID uuid.UUID, from, to time.Time, filter string) ([]CalendarItem, error) {
	rows, err := s.reader.ListByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	participantRows, err := s.reader.ListParticipantEventsByRange(ctx, userID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	rows = mergeCalendarRows(rows, participantRows)

	eventIDs := eventIDsFromSchedules(rows)
	summaries, err := s.reader.ListEventParticipationSummaries(ctx, userID, eventIDs)
	if err != nil {
		return nil, err
	}

	items := make([]CalendarItem, 0, len(rows))
	for _, row := range rows {
		kind := KindSchedule
		if row.Source == schedules.SourceEvent {
			kind = KindEvent
		}
		if !includeForFilter(filter, kind) {
			continue
		}
		item := CalendarItem{
			ID:                  row.ID,
			Kind:                kind,
			Title:               row.Title,
			StartAt:             row.StartAt,
			EndAt:               row.EndAt,
			Location:            row.Location,
			Visibility:          row.Visibility,
			Color:               itemColor(kind, row.Visibility),
			EventID:             row.EventID,
			ParticipantsPreview: []EventParticipantPreview{},
		}
		if row.EventID != nil {
			if summary, ok := summaries[*row.EventID]; ok {
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

func mergeCalendarRows(scheduleRows, participantRows []schedules.Schedule) []schedules.Schedule {
	rows := append([]schedules.Schedule{}, scheduleRows...)
	seenEventIDs := map[int64]bool{}
	for _, row := range scheduleRows {
		if row.Source == schedules.SourceEvent && row.EventID != nil {
			seenEventIDs[*row.EventID] = true
		}
	}
	for _, row := range participantRows {
		if row.EventID == nil {
			continue
		}
		if seenEventIDs[*row.EventID] {
			continue
		}
		seenEventIDs[*row.EventID] = true
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].StartAt.Before(rows[j].StartAt)
	})
	return rows
}

func eventIDsFromSchedules(rows []schedules.Schedule) []int64 {
	seen := map[int64]bool{}
	eventIDs := make([]int64, 0)
	for _, row := range rows {
		if row.Source != schedules.SourceEvent || row.EventID == nil {
			continue
		}
		if seen[*row.EventID] {
			continue
		}
		seen[*row.EventID] = true
		eventIDs = append(eventIDs, *row.EventID)
	}
	return eventIDs
}

func includeForFilter(filter, kind string) bool {
	switch filter {
	case FilterGroups:
		return kind == KindEvent
	default:
		return true
	}
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
