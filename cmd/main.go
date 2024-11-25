package main

import (
	"errors"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"tgbot/internal/bot"
	"tgbot/internal/service"

	"os"

	"tgbot/internal/repository"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := initConfig(); err != nil {
		logrus.Fatalln(err)
	}
	if err := godotenv.Load(); err != nil {
		logrus.Fatalln(errors.New("Error loading .env file"))
	}
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalln(err)
	}
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	token := os.Getenv("TOKEN")
	bot := bot.New(token, services)
	bot.Start()

}
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
