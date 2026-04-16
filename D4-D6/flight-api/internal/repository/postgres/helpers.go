package postgres

import (
	"fmt"
	"strings"
	"time"

	"flightapi/internal/model"
)

var weekdayNames = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func isoWeekday(t time.Time) int {
	w := int(t.Weekday())
	if w == 0 {
		return 7
	}
	return w
}

func pgDaysToAPI(days []int32) []string {
	if len(days) == 0 {
		return nil
	}
	out := make([]string, 0, len(days))
	for _, d := range days {
		if d >= 1 && d <= 7 {
			out = append(out, weekdayNames[d-1])
		}
	}
	return out
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	if d < time.Minute {
		d = time.Minute
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("PT%dH%02dM", h, m)
}

func parsePassengerName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func mapInboundStatus(s string) string {
	switch s {
	case "scheduled":
		return "Scheduled"
	case "delayed":
		return "Delayed"
	case "landed":
		return "Arrived"
	case "cancelled":
		return "Cancelled"
	default:
		return ""
	}
}

func mapOutboundStatus(s string) string {
	switch s {
	case "scheduled":
		return "Scheduled"
	case "delayed":
		return "Delayed"
	case "departed":
		return "Departed"
	case "cancelled":
		return "Cancelled"
	default:
		return ""
	}
}

func clockOnly(t time.Time) time.Time {
	return time.Date(2000, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}

func clipRoutes(routes []model.Route, limit, offset int) []model.Route {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(routes) {
		return []model.Route{}
	}
	end := offset + limit
	if end > len(routes) {
		end = len(routes)
	}
	return routes[offset:end]
}
