package main

import (
	"AvitoTest/internal/db"
	"AvitoTest/internal/jwt"
	"AvitoTest/internal/router"
	"AvitoTest/internal/services/auth"
	"AvitoTest/internal/services/cash"
	"fmt"
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWTConfig    *jwt.Config    `yaml:"jwt"`
	DBConfig     *db.Config     `yaml:"db"`
	RouterConfig *router.Config `yaml:"router"`
}

func readConfig(filename string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := readConfig("./cfg.yml")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("Config read successfully")
	dbi, err := db.New(*cfg.DBConfig)
	if err != nil {
		for i := range 4 {
			slog.Info(fmt.Sprintf("Failed connect to database for the %v time, trying again...", i+1))
			time.Sleep(time.Second * 5)
			dbi, err = db.New(*cfg.DBConfig)
		}
		if err != nil {
			slog.Error("Failed to connect to db after 5 tries: " + err.Error())
			return
		}
	}
	slog.Info("DB connected successfully")
	jwt, err := jwt.New(*cfg.JWTConfig)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("JWT created successfully")
	auth, err := auth.New(dbi, jwt)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	cash, err := cash.New(dbi)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	router := router.New(*cfg.RouterConfig, jwt, auth, cash)
	err = router.Listen()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("Router started successfully, ready to accept requests")
}
