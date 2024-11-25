package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"tgbot"
)

type UserPostgres struct {
	db *sqlx.DB
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{db: db}
}
func (p *UserPostgres) SignIn(login, password string) (tgbot.User, error) {
	var user tgbot.User
	query := fmt.Sprintf("SELECT * FROM users WHERE login = $1 AND password = $2")
	err := p.db.Get(&user, query, login, password)
	if err != nil {
		return tgbot.User{}, err
	}
	return user, nil
}
