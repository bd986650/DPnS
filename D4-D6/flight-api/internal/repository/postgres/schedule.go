package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flightapi/internal/model"
)

func (r *Repository) airportExists(ctx context.Context, code string) (bool, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	var ok bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookings.airports_data WHERE airport_code = $1)`, code,
	).Scan(&ok)
	return ok, err
}

func (r *Repository) GetInboundSchedule(ctx context.Context, airportCode string, q model.ScheduleQuery) ([]model.InboundScheduleEntry, error) {
	ok, err := r.airportExists(ctx, airportCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrAirportNotFound
	}
	airportCode = strings.TrimSpace(strings.ToUpper(airportCode))

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	statusDB := mapInboundStatus(q.Status)

	var sb strings.Builder
	sb.WriteString(`
SELECT r.route_no, r.departure_airport,
  to_char((r.scheduled_time + r.duration)::time, 'HH24:MI:SS') AS arr_time,
  r.days_of_week
FROM bookings.routes r
WHERE r.arrival_airport = $1`)

	args := []interface{}{airportCode}
	argPos := 2

	if q.FlightNumber != "" {
		sb.WriteString(fmt.Sprintf(` AND r.route_no = $%d`, argPos))
		args = append(args, strings.TrimSpace(q.FlightNumber))
		argPos++
	}
	if q.Airline != "" {
		sb.WriteString(fmt.Sprintf(` AND r.route_no LIKE $%d`, argPos))
		args = append(args, strings.TrimSpace(q.Airline)+"%")
		argPos++
	}
	if statusDB != "" {
		sb.WriteString(fmt.Sprintf(` AND EXISTS (
  SELECT 1 FROM bookings.flights f
  WHERE f.route_no = r.route_no AND f.status = $%d
)`, argPos))
		args = append(args, statusDB)
		argPos++
	}

	sb.WriteString(` ORDER BY r.route_no`)

	filterTime := q.From != nil || q.To != nil
	if !filterTime {
		sb.WriteString(fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argPos, argPos+1))
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.InboundScheduleEntry
	for rows.Next() {
		var routeNo, origin string
		var arrTime string
		var days []int32
		if err := rows.Scan(&routeNo, &origin, &arrTime, &days); err != nil {
			return nil, err
		}
		if q.From != nil || q.To != nil {
			t, err := time.Parse("15:04:05", arrTime)
			if err != nil {
				continue
			}
			dayRef := clockOnly(t)
			if q.From != nil && dayRef.Before(clockOnly(*q.From)) {
				continue
			}
			if q.To != nil && dayRef.After(clockOnly(*q.To)) {
				continue
			}
		}
		out = append(out, model.InboundScheduleEntry{
			FlightNo:    strings.TrimSpace(routeNo),
			Origin:      strings.TrimSpace(origin),
			ArrivalTime: arrTime,
			DaysOfWeek:  pgDaysToAPI(days),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if filterTime {
		if offset >= len(out) {
			return []model.InboundScheduleEntry{}, nil
		}
		end := offset + limit
		if end > len(out) {
			end = len(out)
		}
		return out[offset:end], nil
	}
	return out, nil
}

func (r *Repository) GetOutboundSchedule(ctx context.Context, airportCode string, q model.ScheduleQuery) ([]model.OutboundScheduleEntry, error) {
	ok, err := r.airportExists(ctx, airportCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrAirportNotFound
	}
	airportCode = strings.TrimSpace(strings.ToUpper(airportCode))

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	statusDB := mapOutboundStatus(q.Status)

	var sb strings.Builder
	sb.WriteString(`
SELECT r.route_no, r.arrival_airport,
  to_char(r.scheduled_time::time, 'HH24:MI:SS') AS dep_time,
  r.days_of_week
FROM bookings.routes r
WHERE r.departure_airport = $1`)

	args := []interface{}{airportCode}
	argPos := 2

	if q.FlightNumber != "" {
		sb.WriteString(fmt.Sprintf(` AND r.route_no = $%d`, argPos))
		args = append(args, strings.TrimSpace(q.FlightNumber))
		argPos++
	}
	if q.Airline != "" {
		sb.WriteString(fmt.Sprintf(` AND r.route_no LIKE $%d`, argPos))
		args = append(args, strings.TrimSpace(q.Airline)+"%")
		argPos++
	}
	if statusDB != "" {
		sb.WriteString(fmt.Sprintf(` AND EXISTS (
  SELECT 1 FROM bookings.flights f
  WHERE f.route_no = r.route_no AND f.status = $%d
)`, argPos))
		args = append(args, statusDB)
		argPos++
	}

	sb.WriteString(` ORDER BY r.route_no`)

	filterTime := q.From != nil || q.To != nil
	if !filterTime {
		sb.WriteString(fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argPos, argPos+1))
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.OutboundScheduleEntry
	for rows.Next() {
		var routeNo, dest string
		var depTime string
		var days []int32
		if err := rows.Scan(&routeNo, &dest, &depTime, &days); err != nil {
			return nil, err
		}
		if q.From != nil || q.To != nil {
			t, err := time.Parse("15:04:05", depTime)
			if err != nil {
				continue
			}
			dayRef := clockOnly(t)
			if q.From != nil && dayRef.Before(clockOnly(*q.From)) {
				continue
			}
			if q.To != nil && dayRef.After(clockOnly(*q.To)) {
				continue
			}
		}
		out = append(out, model.OutboundScheduleEntry{
			FlightNo:      strings.TrimSpace(routeNo),
			Destination:   strings.TrimSpace(dest),
			DepartureTime: depTime,
			DaysOfWeek:    pgDaysToAPI(days),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if filterTime {
		if offset >= len(out) {
			return []model.OutboundScheduleEntry{}, nil
		}
		end := offset + limit
		if end > len(out) {
			end = len(out)
		}
		return out[offset:end], nil
	}
	return out, nil
}
