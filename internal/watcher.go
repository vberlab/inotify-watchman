package internal

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
	"github.com/vberlabs/inotify-watchman/internal/config"
	"github.com/vberlabs/inotify-watchman/internal/reactor/reactor_core"
)

func Watcher(cfg config.Tracking, exitOnError bool, routineID uint32) error {
	var (
		logger                 *slog.Logger
		watcher                *fsnotify.Watcher
		watcherError           error
		eventQueue             []fsnotify.Event
		reactorsCtx            []*reactor_core.ReactorContext
		fileNamePatternCompile *regexp.Regexp
		reactorCtx             *reactor_core.ReactorContext
		reactorResults         []string
		action                 actions_core.Action
	)
	// Init logger
	logger = slog.With("routineID", routineID)
	// Init fsnotify watcher
	watcher, watcherError = fsnotify.NewWatcher()
	if watcherError != nil {
		return watcherError
	}
	defer watcher.Close()

	watcherError = watcher.Add(cfg.Path)
	if watcherError != nil {
		return watcherError
	}
	if cfg.FileType == "directory" {
		fileNamePatternCompile = regexp.MustCompile(cfg.FileNameFilter)
	}

	for _, reactorCfg := range cfg.Pipeline.Reactor {
		reactorCtx, watcherError = ReactorInit(
			reactorCfg.Name,
			reactorCfg.IgnoreErrors,
			reactorCfg.IgnoreArgsErrors,
			reactorCfg,
		)
		if watcherError != nil {
			logger.Error(fmt.Sprintf("Reactor %s init fail. %s", reactorCfg.Name, watcherError))
			if exitOnError {
				return watcherError
			}
			continue
		}
		reactorsCtx = append(reactorsCtx, reactorCtx)
	}

	for {
		eventQueue, watcherError = watchEvents(watcher, cfg.Pipeline.EventQueueSize, logger, cfg.Events)
		if watcherError != nil {
			if exitOnError {
				return watcherError
			}
			continue
		}

		if cfg.FileType == "directory" {
			eventQueue = filterEventsByFileNames(fileNamePatternCompile, eventQueue, logger)
		}
		if len(eventQueue) == 0 {
			continue
		}

		for _, reactorCtx := range reactorsCtx {
			reactorResults, watcherError = ReactorRun(
				logger,
				reactorCtx,
				map[string]any{"events": eventQueue},
				reactorCtx.ReactorConfig.Args,
			)
			if watcherError != nil {
				logger.Error("Reactor %s condition failed for files %s", reactorCtx.Reactor.ID(), eventQueue)
				if exitOnError {
					return watcherError
				}
			}
			if len(reactorResults) == 0 {
				continue
			}
			logger.Debug(fmt.Sprintf("Reactor %s condition true for files %s", reactorCtx.Reactor.ID(), reactorResults))
			// Iterate over reactor actions
			for _, actionCfg := range reactorCtx.ReactorConfig.Actions {
				action, watcherError = ActionInit(actionCfg.Name, actionCfg)
				if watcherError != nil {
					logger.Error(fmt.Sprintf("Fail to init action %s", actionCfg.Name))
					if exitOnError {
						return watcherError
					}
				}
				_, watcherError = ActionRunWrapper(
					action,
					logger,
					map[string]any{"files": reactorResults},
					actionCfg.Args,
				)
			}
		}
	}
}

func watchEvents(watcher *fsnotify.Watcher, queueLen int, logger *slog.Logger, events []config.Events) ([]fsnotify.Event, error) {
	var (
		recivedEvent fsnotify.Event
		eventQueue   []fsnotify.Event
		err          error
	)
	for {
		select {
		case recivedEvent = <-watcher.Events:
			logger.Debug(fmt.Sprintf("Event received %s %s", recivedEvent.Name, recivedEvent.Op.String()))
			for _, eventType := range events {
				if recivedEvent.Has(fsnotify.Op(eventType)) {
					eventQueue = append(eventQueue, recivedEvent)
				}
			}
			if len(eventQueue) >= queueLen {
				return eventQueue, nil
			}
		case err = <-watcher.Errors:
			logger.Error("Watcher received error", "error", err)
			return nil, err
		}
	}
}

func filterEventsByFileNames(pattern *regexp.Regexp, events []fsnotify.Event, logger *slog.Logger) []fsnotify.Event {
	var matchedEvents []fsnotify.Event
	for _, event := range events {
		if !pattern.MatchString(event.Name) {
			logger.Debug(fmt.Sprintf("File name not mathed %s with %s", event.Name, pattern.String()))
			continue
		}
		matchedEvents = append(matchedEvents, event)
	}
	return matchedEvents
}
