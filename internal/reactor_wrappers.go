package internal

import (
	"fmt"
	"log/slog"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/vberlabs/inotify-watchman/internal/config"
	"github.com/vberlabs/inotify-watchman/internal/reactor/reactor_core"
	"github.com/vberlabs/inotify-watchman/internal/reactor/reactor_registry"
)

func createReactorErrorHandler(logger *slog.Logger) err2.Handler {
	return func(caught error) error {
		logger.Error(fmt.Sprintf("Reactor failed error: %s", caught.Error()))
		return caught
	}
}

func ReactorInit(name string, ignoreErrors bool, ignoreArgsErrors bool, reactorCfg config.Reactor) (*reactor_core.ReactorContext, error) {
	var (
		err        error
		reactor    reactor_core.Reactor
		reactorCtx reactor_core.ReactorContext
	)
	reactor, err = reactor_registry.NewReactor(name)
	if err != nil {
		return nil, err
	}
	reactor.Configure(ignoreErrors, ignoreArgsErrors)
	reactorCtx.Reactor = reactor
	reactorCtx.ReactorConfig = reactorCfg
	return &reactorCtx, nil
}

func ReactorRun(logger *slog.Logger, context *reactor_core.ReactorContext, args ...map[string]any) ([]string, error) {
	var (
		err         error
		reactor     reactor_core.Reactor = context.Reactor
		reactorArgs *reactor_core.ReactorArgs
		result      []string
		errHandler  err2.Handler = createReactorErrorHandler(logger)
	)
	logger.Debug(fmt.Sprintf("Run reactor"))
	logger.Debug(fmt.Sprintf("Prepare arguments"))
	defer err2.Handle(&err, errHandler)
	reactorArgs = try.To1(reactor.PreapreArgs(logger, args...))
	logger.Debug(fmt.Sprintf("Check reactor condition"))
	result = try.To1(reactor.CheckCondition(logger, *reactorArgs))
	return result, err
}
