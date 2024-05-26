package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Audio   Audio      `yaml:"audio"`
	Env     string     `yaml:"env" env-default:"local"`
	Grpc    GrpcConfig `yaml:"grpc"`
	MsGraph MsGraph    `yaml:"msgraph"`
	Tinkoff Tinkoff    `yaml:"tinkoff"`
}

type Audio struct {
	AutomaticPunctuation     bool    `yaml:"automatic_punctuation" env-default:"true"`
	Encoding                 string  `yaml:"encoding" env-default:"LINEAR16"`
	Chans                    int     `yaml:"chans" env-default:"1"`
	LanguageCode             string  `yaml:"language_code" env-default:"ru-RU"`
	MaxAlternatives          int     `yaml:"max_alternatives" env-default:"1"`
	PerformVad               bool    `yaml:"perform_vad" env-default:"true"`
	ProfanityFilter          bool    `yaml:"profanity_filter" env-default:"true"`
	SampleRate               int     `yaml:"sample_rate" env-default:"16000"`
	SilenceDurationThreshold float64 `yaml:"silence_duration_threshold" env-default:"0.6"`
}

type GrpcConfig struct {
	CAFile   string        `yaml:"ca_file"`
	Endpoint string        `yaml:"endpoint" env-default:"api.tinkoff.ai:443"`
	Port     int           `yaml:"port" env-default:"50051"`
	Timeout  time.Duration `yaml:"timeout" env-default:"20s"`
}

type MsGraph struct {
	Token string `yaml:"token"`
}

type Tinkoff struct {
	ApiKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
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
