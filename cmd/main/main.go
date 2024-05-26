package main

import (
	"context"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	"github.com/kanataidarov/tinkoff_voicekit/internal/handler"
	"time"
)

func main() {
	cfg := config.Load()
	log := config.InitLogger(cfg.Env)

	ctx, cancelCtx := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCtx()

	handler.Serve(ctx, cfg, log)
}
