package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (p *PostgresStore) SaveTrade(t Trade) error {
	_, err := p.db.Exec(`INSERT INTO trades (symbol, price, volume) VALUES ($1, $2, $3)`, t.Symbol, t.Price, t.Volume)
	return err
}
