SET search_path = bookings, public;

DROP TABLE IF EXISTS flight_prices_history;
CREATE TABLE flight_prices_history AS
SELECT
  f.flight_id,
  tt.departure_airport,
  tt.arrival_airport,
  s.fare_conditions,
  s.price
FROM flights f
JOIN timetable tt USING (flight_id)
JOIN segments s USING (flight_id);

DROP TABLE IF EXISTS pricing_rules;
CREATE TABLE pricing_rules AS
SELECT
  tt.departure_airport,
  tt.arrival_airport,
  s.fare_conditions,
  percentile_disc(0.5) WITHIN GROUP (ORDER BY s.price) AS rule_price
FROM segments s
JOIN timetable tt USING (flight_id)
GROUP BY
  tt.departure_airport,
  tt.arrival_airport,
  s.fare_conditions;


WITH seg_enriched AS (
  SELECT
    tt.departure_airport,
    tt.arrival_airport,
    s.fare_conditions,
    s.flight_id,
    s.price
  FROM segments s
  JOIN timetable tt USING (flight_id)
),
price_counts AS (
  SELECT
    departure_airport,
    arrival_airport,
    fare_conditions,
    price,
    count(*) AS tickets_at_price
  FROM seg_enriched
  GROUP BY 1,2,3,4
),
group_stats AS (
  SELECT
    departure_airport,
    arrival_airport,
    fare_conditions,
    count(*) AS tickets_cnt,
    count(DISTINCT flight_id) AS flights_cnt,
    avg(price) AS avg_price,
    percentile_cont(0.5) WITHIN GROUP (ORDER BY price) AS median_price
  FROM seg_enriched
  GROUP BY 1,2,3
  HAVING count(DISTINCT price) >= 2
),
price_dist AS (
  SELECT
    departure_airport,
    arrival_airport,
    fare_conditions,
    array_agg(format('%s(%s)', price, tickets_at_price) ORDER BY price) AS ticket_prices
  FROM price_counts
  GROUP BY 1,2,3
)
SELECT
  gs.departure_airport,
  gs.arrival_airport,
  gs.fare_conditions,
  pr.rule_price,
  gs.tickets_cnt,
  gs.flights_cnt,
  gs.avg_price,
  gs.median_price,
  pd.ticket_prices
FROM group_stats gs
JOIN price_dist pd
  ON pd.departure_airport = gs.departure_airport
 AND pd.arrival_airport = gs.arrival_airport
 AND pd.fare_conditions = gs.fare_conditions
LEFT JOIN pricing_rules pr
  ON pr.departure_airport = gs.departure_airport
 AND pr.arrival_airport = gs.arrival_airport
 AND pr.fare_conditions = gs.fare_conditions
ORDER BY gs.tickets_cnt DESC;