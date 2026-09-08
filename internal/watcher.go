package internal

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/vberlabs/inotify-watchman/internal/action"
	"github.com/vberlabs/inotify-watchman/internal/config"
	"github.com/vberlabs/inotify-watchman/internal/reactor"
)

func WatcherHead(cfg config.Tracking, exitOnError bool, routineID uint32) error {
	// Init fsnotify.Wather
	var watcher *fsnotify.Watcher
	var watcherError error
	var eventRecived bool
	var event fsnotify.Event
	var reactors []reactor.ReactorCtx
	var fileNamePetternCompile regexp.Regexp

	logger := slog.With(
		"routine_id", routineID,
	)
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
		fileNamePetternCompile = *regexp.MustCompile(cfg.FileNameFilter)
	}
	// Make reactors
	for _, reactor_cfg := range cfg.Pipeline.Reactor {
		r, err := reactor.NewReactor(reactor_cfg.Name)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: Reactor %s not found in program collection", cfg.Path, reactor_cfg.Name))
		}
		reactors = append(reactors, reactor.ReactorCtx{Reactor: r, ReactorCfgArgs: reactor_cfg.Args})
	}
	// Watch events
	slog.Debug(fmt.Sprintf("%s: Start watching events", cfg.Path))
	for {
		eventRecived, watcherError = watchEvents(watcher, &event, cfg.Events, logger)
		if watcherError != nil {
			if exitOnError {
				return watcherError
			}
		}
		if eventRecived {
			if cfg.FileType == "directory" {
				if !fileNamePetternCompile.MatchString(event.Name) {
					logger.Debug(fmt.Sprintf("%s:File name not mathed %s", cfg.Path, event.Name))
					continue
				}
				logger.Debug(fmt.Sprintf("%s:File name %s matched with %s", cfg.Path, event.Name, cfg.FileNameFilter))
			}
			for _, rCtx := range reactors {
				logger.Debug(fmt.Sprintf("%s:Probe reactor %s", cfg.Path, rCtx.Reactor.ID()))
				_, err := execReactor(&event, &rCtx, logger)
				if err != nil {
					logger.Error(err.Error())
					return err
				}
			}
		}
	}
}

func watchEvents(watcher *fsnotify.Watcher, event *fsnotify.Event, events []config.Events, logger *slog.Logger) (bool, error) {
	for {
		select {
		case recivedEvent := <-watcher.Events:
			logger.Debug(fmt.Sprintf("Event received %s %s", recivedEvent.Name, event.Op.String()))
			for _, eventType := range events {
				if recivedEvent.Has(fsnotify.Op(eventType)) {
					*event = recivedEvent
					return true, nil
				}
			}
		case err := <-watcher.Errors:
			logger.Error("Watcher recived error", "error", err)
			return false, err
		}
	}
}

func execReactor(event *fsnotify.Event, rCtx *reactor.ReactorCtx, logger *slog.Logger) (bool, error) {
	logger.Debug(fmt.Sprintf("Execute reactor %s on %s\n", rCtx.Reactor.ID(), event.Name))
	return rCtx.Reactor.CheckCondition(rCtx.ReactorCfgArgs, event.Name, logger)
}

func executePipelineActions(event *fsnotify.Event, cfgFileType string, cfgPath string, actionsCfg []config.Action, logger *slog.Logger, reactorName string) (map[string]string, error) {
	var act action.Action
	var err error
	var out string
	var results = make(map[string]string)
	var actArgs *action.ActionArgs
	var extArgsPack = map[string]any{
		"config-file-type": cfgFileType,
		"config-path":      cfgPath,
		"reactorName":      reactorName,
		"event-file-name":  event.Name,
	}
	logger.Debug(fmt.Sprintf("Start actions because reactor %s taken true condition", reactorName))
	for _, actCfg := range actionsCfg {
		logger.Debug(fmt.Sprintf("Search %s action in actions collections", actCfg.Name))
		act, err = action.NewAction(actCfg.Name)
		if err != nil {
			return nil, err
		}
		actArgs, err = act.PrepareArgs(logger, extArgsPack, actCfg.Args)
		if err != nil {
			return nil, err
		}
		out, err = act.Execute(*actArgs, logger)
		if err != nil {
			return nil, err
		}
		slog.Debug(fmt.Sprintf("Result output of action %s: %s", actCfg.Name, out))
		results[actCfg.Name] = out
	}
	return results, nil
}
