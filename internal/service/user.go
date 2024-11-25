package service

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"tgbot"
	"tgbot/internal/repository"
)

const (
	salt = ";knmmm3rjoq; 2vr541jdhaDCGV1UE9PED"
)

type UserService struct {
	repo repository.User
}

func NewUserService(repo repository.User) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (u *UserService) SignIn(login, password string) (tgbot.User, error) {
	login = strings.Replace(login, " ", "", -1)
	password = generatePasswordHash(strings.Replace(password, " ", "", -1))
	user, err := u.repo.SignIn(login, password)
	if err != nil {
		return user, err
	}
	return user, nil
}
func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
