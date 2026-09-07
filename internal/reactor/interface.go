package reactor

import "log/slog"

type ReactorCtx struct {
	Reactor        Reactor
	ReactorCfgArgs map[string]any
}

type Reactor interface {
	ID() string
	CheckCondition(cfgArgs map[string]any, fileName string, logger *slog.Logger) (bool, error)
}
