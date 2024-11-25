package repository

import "github.com/jmoiron/sqlx"

type BotPostgres struct {
	db *sqlx.DB
}

func NewBot(db *sqlx.DB) *BotPostgres {
	return &BotPostgres{db: db}
}
