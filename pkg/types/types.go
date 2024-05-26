package types

import (
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	"log/slog"
)

const (
	Group    = "group"
	Meeting  = "meeting"
	OneOnOne = "oneOnOne"
)

type CtxVals struct {
	Config *config.Config
	Logger *slog.Logger
}
