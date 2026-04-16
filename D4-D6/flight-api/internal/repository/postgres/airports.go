package postgres

import (
	"context"
	"fmt"

	"flightapi/internal/cityid"
	"flightapi/internal/ctxlang"
	"flightapi/internal/model"
)

func (r *Repository) ListAirports(ctx context.Context, role string) ([]model.Airport, error) {
	lang := ctxlang.From(ctx)
	apName := fmt.Sprintf("a.airport_name->>'%s'", lang)
	cityName := fmt.Sprintf("a.city->>'%s'", lang)

	var q string
	switch role {
	case "source":
		q = fmt.Sprintf(`
SELECT a.airport_code, %s, a.city->>'en', a.country->>'en', %s
FROM bookings.airports_data a
INNER JOIN bookings.routes rt ON rt.departure_airport = a.airport_code
ORDER BY a.airport_code`, apName, cityName)
	case "destination":
		q = fmt.Sprintf(`
SELECT a.airport_code, %s, a.city->>'en', a.country->>'en', %s
FROM bookings.airports_data a
INNER JOIN bookings.routes rt ON rt.arrival_airport = a.airport_code
ORDER BY a.airport_code`, apName, cityName)
	default:
		q = fmt.Sprintf(`
SELECT a.airport_code, %s, a.city->>'en', a.country->>'en', %s
FROM bookings.airports_data a
ORDER BY a.airport_code`, apName, cityName)
	}

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Airport
	for rows.Next() {
		var code, name, cityEN, countryEN, cityDisp string
		if err := rows.Scan(&code, &name, &cityEN, &countryEN, &cityDisp); err != nil {
			return nil, err
		}
		out = append(out, model.Airport{
			Code:     code,
			Name:     name,
			CityID:   cityid.FromEN(cityEN, countryEN),
			CityName: cityDisp,
		})
	}
	return out, rows.Err()
}

func (r *Repository) ListAirportsByCity(ctx context.Context, cityID string) ([]model.Airport, error) {
	lang := ctxlang.From(ctx)
	apName := fmt.Sprintf("a.airport_name->>'%s'", lang)
	cityName := fmt.Sprintf("a.city->>'%s'", lang)

	dRows, err := r.pool.Query(ctx, `SELECT DISTINCT city->>'en', country->>'en' FROM bookings.airports_data`)
	if err != nil {
		return nil, err
	}
	var cityEN, countryEN string
	found := false
	for dRows.Next() {
		var ce, co string
		if err := dRows.Scan(&ce, &co); err != nil {
			dRows.Close()
			return nil, err
		}
		if cityid.FromEN(ce, co) == cityID {
			cityEN, countryEN = ce, co
			found = true
			break
		}
	}
	dRows.Close()
	if err := dRows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, model.ErrCityNotFound
	}

	q := fmt.Sprintf(`
SELECT a.airport_code, %s, %s
FROM bookings.airports_data a
WHERE a.city->>'en' = $1 AND a.country->>'en' = $2
ORDER BY a.airport_code`, apName, cityName)

	rows, err := r.pool.Query(ctx, q, cityEN, countryEN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Airport
	for rows.Next() {
		var code, name, cityDisp string
		if err := rows.Scan(&code, &name, &cityDisp); err != nil {
			return nil, err
		}
		out = append(out, model.Airport{
			Code:     code,
			Name:     name,
			CityID:   cityID,
			CityName: cityDisp,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, model.ErrCityNotFound
	}
	return out, nil
}
