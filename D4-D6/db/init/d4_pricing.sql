SET search_path = bookings, public;

DROP TABLE IF EXISTS flight_prices_history;
CREATE TABLE flight_prices_history AS
SELECT
  f.flight_id,
  tt.departure_airport,
  tt.arrival_airport,
  s.fare_conditions,
  percentile_disc(0.5) WITHIN GROUP (ORDER BY s.price) AS flight_price
FROM flights f
JOIN timetable tt USING (flight_id)
JOIN segments s  USING (flight_id)
GROUP BY
  f.flight_id,
  tt.departure_airport,
  tt.arrival_airport,
  s.fare_conditions;

DROP TABLE IF EXISTS pricing_rules;
CREATE TABLE pricing_rules AS
SELECT
  departure_airport,
  arrival_airport,
  fare_conditions,
  percentile_disc(0.5) WITHIN GROUP (ORDER BY flight_price) AS rule_price
FROM flight_prices_history
GROUP BY 1,2,3;