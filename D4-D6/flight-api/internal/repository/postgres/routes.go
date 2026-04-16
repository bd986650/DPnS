package postgres

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"flightapi/internal/helper"
	"flightapi/internal/model"
)

type routeEdge struct {
	RouteNo  string
	Dep, Arr string
	Airplane string
	DurSec   int64
}

type flightRow struct {
	ID        int64
	RouteNo   string
	Status    string
	Dep, Arr  time.Time
}

func sameLocalDay(t time.Time, day time.Time, loc *time.Location) bool {
	lt := t.In(loc)
	d := day.In(loc)
	return lt.Year() == d.Year() && lt.Month() == d.Month() && lt.Day() == d.Day()
}

func (r *Repository) fareClasses(ctx context.Context, airplane string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT fare_conditions FROM bookings.seats WHERE airplane_code = $1 ORDER BY fare_conditions`,
		strings.TrimSpace(airplane),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var fc string
		if err := rows.Scan(&fc); err != nil {
			return nil, err
		}
		out = append(out, fc)
	}
	return out, rows.Err()
}

func (r *Repository) FindRoutes(
	ctx context.Context,
	originAirports []string,
	destAirports []string,
	p model.RouteSearchParams,
) ([]model.Route, error) {
	loc := time.UTC
	if p.Timezone != "" {
		if l, err := time.LoadLocation(p.Timezone); err == nil {
			loc = l
		}
	}

	day := time.Date(
		p.DepartureDate.Year(), p.DepartureDate.Month(), p.DepartureDate.Day(),
		0, 0, 0, 0, loc,
	)
	weekday := isoWeekday(day)
	startUTC := day.UTC()
	endUTC := day.Add(24 * time.Hour).UTC()

	rows, err := r.pool.Query(ctx, `
SELECT route_no, trim(departure_airport), trim(arrival_airport), trim(airplane_code),
       EXTRACT(EPOCH FROM duration)::bigint
FROM bookings.routes
WHERE validity && tstzrange($1::timestamptz, $2::timestamptz)
  AND $3::int = ANY(days_of_week)
`, startUTC, endUTC, weekday)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []routeEdge
	for rows.Next() {
		var e routeEdge
		if err := rows.Scan(&e.RouteNo, &e.Dep, &e.Arr, &e.Airplane, &e.DurSec); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	winStart := startUTC.Add(-12 * time.Hour)
	winEnd := endUTC.Add(36 * time.Hour)

	fRows, err := r.pool.Query(ctx, `
SELECT flight_id, route_no, status, scheduled_departure, scheduled_arrival
FROM bookings.flights
WHERE scheduled_departure >= $1 AND scheduled_departure < $2
  AND status NOT IN ('Cancelled')
`, winStart, winEnd)
	if err != nil {
		return nil, err
	}
	defer fRows.Close()

	flightsByRoute := make(map[string][]flightRow)
	for fRows.Next() {
		var f flightRow
		if err := fRows.Scan(&f.ID, &f.RouteNo, &f.Status, &f.Dep, &f.Arr); err != nil {
			return nil, err
		}
		rn := strings.TrimSpace(f.RouteNo)
		flightsByRoute[rn] = append(flightsByRoute[rn], f)
	}
	if err := fRows.Err(); err != nil {
		return nil, err
	}
	for k := range flightsByRoute {
		sort.Slice(flightsByRoute[k], func(i, j int) bool {
			return flightsByRoute[k][i].Dep.Before(flightsByRoute[k][j].Dep)
		})
	}

	originSet := make(map[string]struct{}, len(originAirports))
	for _, o := range originAirports {
		originSet[strings.TrimSpace(strings.ToUpper(o))] = struct{}{}
	}
	destSet := make(map[string]struct{}, len(destAirports))
	for _, d := range destAirports {
		destSet[strings.TrimSpace(strings.ToUpper(d))] = struct{}{}
	}

	maxHops := 4
	if p.MaxConnections >= 0 {
		maxHops = p.MaxConnections + 1
	}
	if maxHops > 4 {
		maxHops = 4
	}

	byDep := make(map[string][]routeEdge)
	for _, e := range edges {
		byDep[e.Dep] = append(byDep[e.Dep], e)
	}

	var paths [][]routeEdge
	var queue [][]routeEdge
	seen := 0
	const maxQueue = 8000

	for _, o := range originAirports {
		o = strings.TrimSpace(strings.ToUpper(o))
		for _, e := range byDep[o] {
			if _, ok := originSet[e.Dep]; ok {
				queue = append(queue, []routeEdge{e})
			}
		}
	}

	for len(queue) > 0 && seen < maxQueue {
		path := queue[0]
		queue = queue[1:]
		seen++

		last := path[len(path)-1]
		if _, ok := destSet[last.Arr]; ok {
			paths = append(paths, append([]routeEdge(nil), path...))
		}

		if len(path) >= maxHops {
			continue
		}

		visited := make(map[string]struct{}, len(path)+2)
		for _, e := range path {
			visited[e.Dep] = struct{}{}
			visited[e.Arr] = struct{}{}
		}

		for _, next := range byDep[last.Arr] {
			if _, dup := visited[next.Arr]; dup {
				continue
			}
			queue = append(queue, append(append([]routeEdge(nil), path...), next))
		}
	}

	var routes []model.Route
	for _, path := range paths {
		flights, ok := pickFlightsForPath(path, flightsByRoute, day, loc)
		if !ok || len(flights) != len(path) {
			continue
		}

		classCache := make(map[string][]string)
		for _, e := range path {
			if _, ok := classCache[e.Airplane]; ok {
				continue
			}
			cl, err := r.fareClasses(ctx, e.Airplane)
			if err != nil {
				return nil, err
			}
			classCache[e.Airplane] = cl
		}
		if p.BookingClass != "" {
			okClass := true
			for _, e := range path {
				if !helper.ContainsString(classCache[e.Airplane], p.BookingClass) {
					okClass = false
					break
				}
			}
			if !okClass {
				continue
			}
		}

		var legs []model.FlightLeg
		var total time.Duration
		for i, e := range path {
			f := flights[i]
			cl := classCache[e.Airplane]
			legDur := f.Arr.Sub(f.Dep)
			if legDur < time.Minute {
				legDur = time.Duration(e.DurSec) * time.Second
			}
			total += legDur

			legs = append(legs, model.FlightLeg{
				FlightNo:           strings.TrimSpace(e.RouteNo),
				OriginAirport:      e.Dep,
				DestinationAirport: e.Arr,
				DepartureDateTime:  f.Dep,
				ArrivalDateTime:    f.Arr,
				AvailableClasses:   append([]string(nil), cl...),
			})
		}

		ids := make([]string, len(flights))
		for i := range flights {
			ids[i] = strconv.FormatInt(flights[i].ID, 10)
		}
		routeID := "F" + strings.Join(ids, "-")

		routes = append(routes, model.Route{
			RouteID:       routeID,
			Connections:   len(path) - 1,
			TotalDuration: formatDuration(total),
			Legs:          legs,
		})
	}

	sort.Slice(routes, func(i, j int) bool {
		if len(routes[i].Legs) == 0 || len(routes[j].Legs) == 0 {
			return false
		}
		return routes[i].Legs[0].DepartureDateTime.Before(routes[j].Legs[0].DepartureDateTime)
	})

	return clipRoutes(routes, p.Limit, p.Offset), nil
}

func pickFlightsForPath(path []routeEdge, flightsByRoute map[string][]flightRow, day time.Time, loc *time.Location) ([]flightRow, bool) {
	out := make([]flightRow, 0, len(path))
	var prevArrival time.Time
	for i, e := range path {
		cands := flightsByRoute[strings.TrimSpace(e.RouteNo)]
		var best *flightRow
		for j := range cands {
			f := &cands[j]
			if f.Status == "Cancelled" {
				continue
			}
			if i == 0 {
				if !sameLocalDay(f.Dep, day, loc) {
					continue
				}
			} else {
				if f.Dep.Before(prevArrival.Add(25 * time.Minute)) {
					continue
				}
			}
			if best == nil || f.Dep.Before(best.Dep) {
				best = f
			}
		}
		if best == nil {
			return nil, false
		}
		out = append(out, *best)
		prevArrival = best.Arr
	}
	return out, true
}
