package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabasURI           string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Key                  string
	MigrationDir         string
}

func InitConfig() Config {
	var cfg = Config{
		RunAddress:           "http://localhost:8080",
		DatabasURI:           "",
		AccrualSystemAddress: "",
		Key:                  "]X5uhRFCcd4gU",
		MigrationDir:         "./migrations",
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "Run address")
	flag.StringVar(&cfg.DatabasURI, "d", cfg.DatabasURI, "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.Parse()

	return cfg
}
