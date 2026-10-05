package internal

import (
	"fmt"
	"log/slog"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_registry"
	"github.com/vberlabs/inotify-watchman/internal/config"
)

func CreateActionErrHandler(logger *slog.Logger) err2.Handler {
	return func(caught error) error {
		logger.Error(fmt.Sprintf("Reactor failed error: %s", caught.Error()))
		return caught
	}
}
func ActionInit(name string, cfg config.Action) (actions_core.Action, error) {
	var (
		action actions_core.Action
		err    error
	)
	action, err = actions_registry.NewAction(name)
	if err != nil {
		return nil, err
	}
	action.Configure(cfg.IgnoreErrors, cfg.IgnoreArgsErrors)
	return action, nil
}

func ActionRunWrapper(action actions_core.Action, logger *slog.Logger, args ...map[string]any) ([]string, error) {
	var (
		err           error
		actionArgs    *actions_core.ActionArgs
		actionResults []string
		errHandler    err2.Handler = CreateActionErrHandler(logger)
	)
	defer err2.Handle(&err, errHandler)
	actionArgs = try.To1(action.PreapreArgs(logger, args...))
	actionResults = try.To1(action.RunAction(*actionArgs, logger))
	return actionResults, nil
}
