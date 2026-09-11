package reactor

import (
	"log/slog"

	"github.com/vberlabs/inotify-watchman/internal/config"
)

type ReactorCtx struct {
	Reactor           Reactor
	ReactorCfgArgs    map[string]any
	ReactorActionsCfg []config.Action
}

type Reactor interface {
	ID() string
	CheckCondition(cfgArgs map[string]any, fileName string, logger *slog.Logger) (bool, error)
}
