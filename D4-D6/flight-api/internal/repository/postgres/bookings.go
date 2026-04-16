package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"flightapi/internal/helper"
	"flightapi/internal/model"
)

func parseRouteFlightIDs(routeID string) ([]int64, error) {
	s := strings.TrimPrefix(strings.TrimSpace(routeID), "F")
	if s == strings.TrimSpace(routeID) {
		return nil, model.ErrRouteNotFound
	}
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return nil, model.ErrRouteNotFound
	}
	var ids []int64
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, model.ErrRouteNotFound
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func randomBookRef() (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b.WriteByte(chars[n.Int64()])
	}
	return b.String(), nil
}

func (r *Repository) CreateBooking(ctx context.Context, req model.BookingRequest) (model.Booking, error) {
	ids, err := parseRouteFlightIDs(req.RouteID)
	if err != nil {
		return model.Booking{}, err
	}

	type fRow struct {
		ID       int64
		RouteNo  string
		Dep, Arr time.Time
		Plane    string
		DepAP    string
		ArrAP    string
	}
	flights := make([]fRow, 0, len(ids))
	for _, id := range ids {
		var fr fRow
		err := r.pool.QueryRow(ctx, `
SELECT f.flight_id, trim(f.route_no), f.scheduled_departure, f.scheduled_arrival,
       trim(r.airplane_code), trim(r.departure_airport), trim(r.arrival_airport)
FROM bookings.flights f
JOIN bookings.routes r ON r.route_no = f.route_no AND r.validity @> f.scheduled_departure
WHERE f.flight_id = $1
`, id).Scan(&fr.ID, &fr.RouteNo, &fr.Dep, &fr.Arr, &fr.Plane, &fr.DepAP, &fr.ArrAP)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return model.Booking{}, model.ErrRouteNotFound
			}
			return model.Booking{}, err
		}
		flights = append(flights, fr)
	}

	for i := 1; i < len(flights); i++ {
		if flights[i-1].ArrAP != flights[i].DepAP {
			return model.Booking{}, model.ErrRouteNotFound
		}
	}

	for _, fr := range flights {
		ok, err := r.hasFreeSeat(ctx, fr.ID, fr.Plane, req.BookingClass)
		if err != nil {
			return model.Booking{}, err
		}
		if !ok {
			return model.Booking{}, model.ErrNoSeatsAvailable
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Booking{}, err
	}
	defer tx.Rollback(ctx)

	var bookRef string
	for attempt := 0; attempt < 8; attempt++ {
		br, err := randomBookRef()
		if err != nil {
			return model.Booking{}, err
		}
		var exists bool
		_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings.bookings WHERE book_ref = $1)`, br).Scan(&exists)
		if !exists {
			bookRef = br
			break
		}
	}
	if bookRef == "" {
		return model.Booking{}, fmt.Errorf("could not allocate book_ref")
	}

	var total float64
	for _, fr := range flights {
		var segPrice float64
		err := tx.QueryRow(ctx, `
SELECT COALESCE(AVG(price), 0)
FROM bookings.segments
WHERE flight_id = $1 AND fare_conditions = $2
`, fr.ID, req.BookingClass).Scan(&segPrice)
		if err != nil {
			return model.Booking{}, err
		}
		if segPrice < 1 {
			segPrice = 500
		}
		total += segPrice
	}
	if total < 1 {
		total = float64(len(flights)) * 500
	}

	_, err = tx.Exec(ctx, `
INSERT INTO bookings.bookings (book_ref, book_date, total_amount)
VALUES ($1, now(), $2)
`, bookRef, total)
	if err != nil {
		return model.Booking{}, err
	}

	var maxTicket int64
	_ = tx.QueryRow(ctx, `SELECT COALESCE(MAX(ticket_no::bigint), 0) FROM bookings.tickets`).Scan(&maxTicket)
	ticketNo := fmt.Sprintf("%013d", maxTicket+1)

	passengerName := strings.TrimSpace(req.Passenger.FirstName + " " + req.Passenger.LastName)
	_, err = tx.Exec(ctx, `
INSERT INTO bookings.tickets (ticket_no, book_ref, passenger_id, passenger_name, outbound)
VALUES ($1, $2, $3, $4, true)
`, ticketNo, bookRef, req.Passenger.DocumentNumber, passengerName)
	if err != nil {
		return model.Booking{}, err
	}

	for _, fr := range flights {
		var segPrice float64
		err := tx.QueryRow(ctx, `
SELECT COALESCE(AVG(price), 500)
FROM bookings.segments
WHERE flight_id = $1 AND fare_conditions = $2
`, fr.ID, req.BookingClass).Scan(&segPrice)
		if err != nil {
			return model.Booking{}, err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO bookings.segments (ticket_no, flight_id, fare_conditions, price)
VALUES ($1, $2, $3, $4)
`, ticketNo, fr.ID, req.BookingClass, segPrice)
		if err != nil {
			return model.Booking{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Booking{}, err
	}

	return r.GetBooking(ctx, bookRef)
}

func (r *Repository) hasFreeSeat(ctx context.Context, flightID int64, airplane, fare string) (bool, error) {
	var cap int
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM bookings.seats WHERE airplane_code = $1 AND fare_conditions = $2
`, strings.TrimSpace(airplane), fare).Scan(&cap)
	if err != nil {
		return false, err
	}
	if cap == 0 {
		return false, nil
	}
	var used int
	err = r.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM bookings.segments WHERE flight_id = $1 AND fare_conditions = $2
`, flightID, fare).Scan(&used)
	if err != nil {
		return false, err
	}
	return used < cap, nil
}

func (r *Repository) GetBooking(ctx context.Context, bookingID string) (model.Booking, error) {
	var bookDate time.Time
	var ticketNo, passengerName, passengerID string
	err := r.pool.QueryRow(ctx, `
SELECT b.book_date, t.ticket_no, t.passenger_name, t.passenger_id
FROM bookings.bookings b
JOIN bookings.tickets t ON t.book_ref = b.book_ref
WHERE b.book_ref = $1
`, strings.TrimSpace(bookingID)).Scan(&bookDate, &ticketNo, &passengerName, &passengerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Booking{}, model.ErrBookingNotFound
		}
		return model.Booking{}, err
	}

	rows, err := r.pool.Query(ctx, `
SELECT f.flight_id, trim(f.route_no), f.scheduled_departure, f.scheduled_arrival,
       trim(r.departure_airport), trim(r.arrival_airport), s.fare_conditions, trim(r.airplane_code)
FROM bookings.segments s
JOIN bookings.flights f ON f.flight_id = s.flight_id
JOIN bookings.routes r ON r.route_no = f.route_no AND r.validity @> f.scheduled_departure
WHERE s.ticket_no = $1
ORDER BY f.scheduled_departure
`, ticketNo)
	if err != nil {
		return model.Booking{}, err
	}
	defer rows.Close()

	var legs []model.FlightLeg
	var bookingClass string
	var routeParts []string
	for rows.Next() {
		var fid int64
		var routeNo string
		var dep, arr time.Time
		var depAP, arrAP, fc, plane string
		if err := rows.Scan(&fid, &routeNo, &dep, &arr, &depAP, &arrAP, &fc, &plane); err != nil {
			return model.Booking{}, err
		}
		if bookingClass == "" {
			bookingClass = fc
		}
		routeParts = append(routeParts, strconv.FormatInt(fid, 10))
		cl, err := r.fareClasses(ctx, plane)
		if err != nil {
			return model.Booking{}, err
		}
		legs = append(legs, model.FlightLeg{
			FlightNo:           strings.TrimSpace(routeNo),
			OriginAirport:      depAP,
			DestinationAirport: arrAP,
			DepartureDateTime:  dep,
			ArrivalDateTime:    arr,
			AvailableClasses:   cl,
		})
	}
	if err := rows.Err(); err != nil {
		return model.Booking{}, err
	}

	routeID := "F" + strings.Join(routeParts, "-")
	fn, ln := parsePassengerName(passengerName)
	return model.Booking{
		BookingID:    bookingID,
		Status:       "confirmed",
		RouteID:      routeID,
		BookingClass: bookingClass,
		Passenger: model.Passenger{
			FirstName:      fn,
			LastName:       ln,
			DateOfBirth:    "",
			DocumentNumber: passengerID,
		},
		Legs:      legs,
		CreatedAt: bookDate,
	}, nil
}

func (r *Repository) CheckIn(ctx context.Context, bookingID string, req model.CheckInRequest) (model.BoardingPass, error) {
	b, err := r.GetBooking(ctx, bookingID)
	if err != nil {
		return model.BoardingPass{}, err
	}

	var ticketNo string
	err = r.pool.QueryRow(ctx, `SELECT ticket_no FROM bookings.tickets WHERE book_ref = $1 LIMIT 1`, bookingID).Scan(&ticketNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BoardingPass{}, model.ErrBookingNotFound
		}
		return model.BoardingPass{}, err
	}

	var flightID int64
	var routeNo string
	err = r.pool.QueryRow(ctx, `
SELECT f.flight_id, trim(f.route_no)
FROM bookings.segments s
JOIN bookings.flights f ON f.flight_id = s.flight_id
WHERE s.ticket_no = $1 AND trim(f.route_no) = trim($2)
LIMIT 1
`, ticketNo, req.FlightNo).Scan(&flightID, &routeNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BoardingPass{}, model.ErrCheckInNotAvailable
		}
		return model.BoardingPass{}, err
	}

	var already bool
	err = r.pool.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM bookings.boarding_passes WHERE ticket_no = $1 AND flight_id = $2)
`, ticketNo, flightID).Scan(&already)
	if err != nil {
		return model.BoardingPass{}, err
	}
	if already {
		return model.BoardingPass{}, model.ErrCheckInNotAvailable
	}

	var plane, fare string
	err = r.pool.QueryRow(ctx, `
SELECT trim(r.airplane_code), trim(s.fare_conditions)
FROM bookings.segments s
JOIN bookings.flights f ON f.flight_id = s.flight_id
JOIN bookings.routes r ON r.route_no = f.route_no AND r.validity @> f.scheduled_departure
WHERE s.ticket_no = $1 AND f.flight_id = $2
`, ticketNo, flightID).Scan(&plane, &fare)
	if err != nil {
		return model.BoardingPass{}, err
	}

	seat := helper.PickSeat(req.SeatPreference)
	var picked string
	err = r.pool.QueryRow(ctx, `
SELECT s.seat_no FROM bookings.seats s
WHERE s.airplane_code = $1 AND s.fare_conditions = $2
  AND s.seat_no NOT IN (SELECT bp.seat_no FROM bookings.boarding_passes bp WHERE bp.flight_id = $3)
ORDER BY s.seat_no LIMIT 1
`, plane, fare, flightID).Scan(&picked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			picked = seat
		} else {
			return model.BoardingPass{}, err
		}
	} else {
		seat = picked
	}

	var boardingNo int
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(boarding_no), 0) FROM bookings.boarding_passes WHERE flight_id = $1`, flightID).Scan(&boardingNo)
	boardingNo++

	bt := time.Now().UTC().Add(40 * time.Minute)
	_, err = r.pool.Exec(ctx, `
INSERT INTO boarding_passes (ticket_no, flight_id, seat_no, boarding_no, boarding_time)
VALUES ($1, $2, $3, $4, $5)
`, ticketNo, flightID, seat, boardingNo, bt)
	if err != nil {
		return model.BoardingPass{}, err
	}

	passengerName := strings.TrimSpace(b.Passenger.FirstName + " " + b.Passenger.LastName)
	return model.BoardingPass{
		BookingID:     bookingID,
		FlightNo:      strings.TrimSpace(routeNo),
		PassengerName: passengerName,
		Seat:          seat,
		Gate:          fmt.Sprintf("%c%d", 'A'+(int(flightID)%3), 10+(int(flightID)%20)),
		BoardingTime:  bt,
		Barcode:       fmt.Sprintf("BP-%s-%s-%d", bookingID, routeNo, flightID),
	}, nil
}
