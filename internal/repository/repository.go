package repository

import (
	"github.com/jmoiron/sqlx"
	"tgbot"
)

type Bot interface {
}
type User interface {
	SignIn(string, string) (tgbot.User, error)
}
type Repository struct {
	Bot
	User
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Bot:  NewBot(db),
		User: NewUserPostgres(db),
	}
}
