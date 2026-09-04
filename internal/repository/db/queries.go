package db

import "database/sql"

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) DB() *sql.DB {
	return q.db
}
