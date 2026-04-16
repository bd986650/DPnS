package postgres

import "github.com/jackc/pgx/v5/pgxpool"

// Repository читает данные из PostgreSQL (схема bookings, БД demo).
type Repository struct {
	pool *pgxpool.Pool
}

// New создаёт репозиторий.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
