package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Env  string     `yaml:"env" env-default:"local"`
	Grpc GrpcConfig `yaml:"grpc"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port" env-default:"50051"`
	Timeout time.Duration `yaml:"timeout" env-default:"20s"`
}

func Load() *Config {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		panic("CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic("File at CONFIG_PATH does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic("Could not read config. " + err.Error())
	}

	return &cfg
}

func InitLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "dev", "local":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
