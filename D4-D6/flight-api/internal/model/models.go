package model

import (
	"context"
	"errors"
	"time"
)

type City struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type Airport struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	CityID   string `json:"cityId"`
	CityName string `json:"cityName"`
}

type InboundScheduleEntry struct {
	FlightNo    string   `json:"flightNo"`
	Origin      string   `json:"origin"`
	ArrivalTime string   `json:"arrivalTime"`
	DaysOfWeek  []string `json:"daysOfWeek"`
}

type OutboundScheduleEntry struct {
	FlightNo      string   `json:"flightNo"`
	Destination   string   `json:"destination"`
	DepartureTime string   `json:"departureTime"`
	DaysOfWeek    []string `json:"daysOfWeek"`
}

type FlightLeg struct {
	FlightNo           string    `json:"flightNo"`
	OriginAirport      string    `json:"originAirport"`
	DestinationAirport string    `json:"destinationAirport"`
	DepartureDateTime  time.Time `json:"departureDateTime"`
	ArrivalDateTime    time.Time `json:"arrivalDateTime"`
	AvailableClasses   []string  `json:"availableClasses"`
}

type Route struct {
	RouteID       string      `json:"routeId"`
	Connections   int         `json:"connections"`
	TotalDuration string      `json:"totalDuration"`
	Legs          []FlightLeg `json:"legs"`
}

type Passenger struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	DateOfBirth    string `json:"dateOfBirth"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	DocumentNumber string `json:"documentNumber"`
}

type BookingRequest struct {
	RouteID      string    `json:"routeId"`
	BookingClass string    `json:"bookingClass"`
	Passenger    Passenger `json:"passenger"`
}

type Booking struct {
	BookingID    string      `json:"bookingId"`
	Status       string      `json:"status"`
	RouteID      string      `json:"routeId"`
	BookingClass string      `json:"bookingClass"`
	Passenger    Passenger   `json:"passenger"`
	Legs         []FlightLeg `json:"legs"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type CheckInRequest struct {
	FlightNo       string `json:"flightNo"`
	SeatPreference string `json:"seatPreference,omitempty"`
}

type ScheduleQuery struct {
	From         *time.Time
	To           *time.Time
	Timezone     string
	Status       string
	Airline      string
	FlightNumber string
	Limit        int
	Offset       int
}

type RouteSearchParams struct {
	DepartureDate  time.Time
	BookingClass   string
	MaxConnections int
	Timezone       string
	Limit, Offset  int
}

type BoardingPass struct {
	BookingID     string    `json:"bookingId"`
	FlightNo      string    `json:"flightNo"`
	PassengerName string    `json:"passengerName"`
	Seat          string    `json:"seat"`
	Gate          string    `json:"gate"`
	BoardingTime  time.Time `json:"boardingTime"`
	Barcode       string    `json:"barcode"`
}

var (
	ErrCityNotFound        = errors.New("city not found")
	ErrAirportNotFound     = errors.New("airport not found")
	ErrRouteNotFound       = errors.New("route not found")
	ErrBookingNotFound     = errors.New("booking not found")
	ErrNoSeatsAvailable    = errors.New("no seats available")
	ErrPricingRuleNotFound = errors.New("pricing rule not found")
	ErrCheckInNotAvailable = errors.New("check-in not available for this flight")
)

type FlightRepository interface {
	ListCities(ctx context.Context, role string) ([]City, error)
	ListAirports(ctx context.Context, role string) ([]Airport, error)
	ListAirportsByCity(ctx context.Context, cityID string) ([]Airport, error)

	GetInboundSchedule(ctx context.Context, airportCode string, q ScheduleQuery) ([]InboundScheduleEntry, error)
	GetOutboundSchedule(ctx context.Context, airportCode string, q ScheduleQuery) ([]OutboundScheduleEntry, error)

	FindRoutes(
		ctx context.Context,
		originAirports []string,
		destAirports []string,
		p RouteSearchParams,
	) ([]Route, error)

	CreateBooking(ctx context.Context, req BookingRequest) (Booking, error)
	GetBooking(ctx context.Context, bookingID string) (Booking, error)
	CheckIn(ctx context.Context, bookingID string, req CheckInRequest) (BoardingPass, error)
}
