package main

import (
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	"github.com/kanataidarov/tinkoff_voicekit/internal/handler"
)

func main() {
	cfg := config.Load()
	log := config.InitLogger(cfg.Env)

	handler.Serve(cfg, log)
}
