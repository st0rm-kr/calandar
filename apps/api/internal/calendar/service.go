package calendar

import (
	"context"
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

type CalendarItem struct {
	ID          int64      `json:"id"`
	Kind        string     `json:"kind"`
	Title       string     `json:"title"`
	StartAt     time.Time  `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	Location    *string    `json:"location"`
	Visibility  string     `json:"visibility"`
	Color       string     `json:"color"`
	HasConflict bool       `json:"has_conflict"`
	EventID     *int64     `json:"event_id"`
}

type Service struct {
	reader ScheduleReader
}

func NewService(reader ScheduleReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) Month(ctx context.Context, userID uuid.UUID, from, to time.Time, filter string) ([]CalendarItem, error) {
	rows, err := s.reader.ListByRange(ctx, userID, from.UTC(), to.UTC())
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
		items = append(items, CalendarItem{
			ID:         row.ID,
			Kind:       kind,
			Title:      row.Title,
			StartAt:    row.StartAt,
			EndAt:      row.EndAt,
			Location:   row.Location,
			Visibility: row.Visibility,
			Color:      itemColor(kind, row.Visibility),
			EventID:    row.EventID,
		})
	}

	markConflicts(items)
	return items, nil
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
