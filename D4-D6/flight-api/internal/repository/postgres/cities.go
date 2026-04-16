package postgres

import (
	"context"
	"fmt"

	"flightapi/internal/cityid"
	"flightapi/internal/ctxlang"
	"flightapi/internal/model"
)

func (r *Repository) ListCities(ctx context.Context, role string) ([]model.City, error) {
	lang := ctxlang.From(ctx)
	langCol := fmt.Sprintf("city->>'%s'", lang)
	countryCol := fmt.Sprintf("country->>'%s'", lang)

	var q string
	switch role {
	case "source":
		q = fmt.Sprintf(`
SELECT DISTINCT ON (a.city->>'en', a.country->>'en')
  a.city->>'en', a.country->>'en', %s, %s
FROM bookings.airports_data a
INNER JOIN bookings.routes rt ON rt.departure_airport = a.airport_code
ORDER BY a.city->>'en', a.country->>'en'`, langCol, countryCol)
	case "destination":
		q = fmt.Sprintf(`
SELECT DISTINCT ON (a.city->>'en', a.country->>'en')
  a.city->>'en', a.country->>'en', %s, %s
FROM bookings.airports_data a
INNER JOIN bookings.routes rt ON rt.arrival_airport = a.airport_code
ORDER BY a.city->>'en', a.country->>'en'`, langCol, countryCol)
	default:
		q = fmt.Sprintf(`
SELECT DISTINCT ON (a.city->>'en', a.country->>'en')
  a.city->>'en', a.country->>'en', %s, %s
FROM bookings.airports_data a
ORDER BY a.city->>'en', a.country->>'en'`, langCol, countryCol)
	}

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.City
	for rows.Next() {
		var cityEN, countryEN, name, country string
		if err := rows.Scan(&cityEN, &countryEN, &name, &country); err != nil {
			return nil, err
		}
		out = append(out, model.City{
			ID:      cityid.FromEN(cityEN, countryEN),
			Name:    name,
			Country: country,
		})
	}
	return out, rows.Err()
}
