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
	var eventsQueue []fsnotify.Event
	var eventsReactorCondition []fsnotify.Event
	var reactors []reactor.ReactorCtx
	var fileNamePetternCompile regexp.Regexp
	// Add ID to logger
	logger := slog.With(
		"routine_id", routineID,
	)
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
		fileNamePetternCompile = *regexp.MustCompile(cfg.FileNameFilter)
	}

	// Make reactors
	for _, reactor_cfg := range cfg.Pipeline.Reactor {
		r, err := reactor.NewReactor(reactor_cfg.Name)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: Reactor %s not found in program collection", cfg.Path, reactor_cfg.Name))
		}
		reactors = append(reactors, reactor.ReactorCtx{Reactor: r, ReactorCfgArgs: reactor_cfg.Args, ReactorActionsCfg: reactor_cfg.Actions})
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
		// Add event to qeue
		if eventRecived {
			if cfg.FileType == "directory" {
				if !fileNamePetternCompile.MatchString(event.Name) {
					logger.Debug(fmt.Sprintf("%s:File name not mathed %s", cfg.Path, event.Name))
					continue
				}
				logger.Debug(fmt.Sprintf("%s:File name %s matched with %s", cfg.Path, event.Name, cfg.FileNameFilter))
			}
			eventsQueue = append(eventsQueue, event)
		}
		logger.Debug(fmt.Sprintf("Fsnotify Queue len is %d. Config queue len is %d", len(eventsQueue), cfg.Pipeline.EventQueueSize))
		if len(eventsQueue) >= int(cfg.Pipeline.EventQueueSize) {
			for _, reactorItem := range reactors {
				for _, eventQItem := range eventsQueue {
					logger.Debug(fmt.Sprintf("Probe reactor %s condition for %s", reactorItem.Reactor.ID(), eventQItem.Name))
					rCondition, err := execReactor(&eventQItem, &reactorItem, logger)
					if err != nil {
						logger.Error(err.Error())
					}
					if !rCondition {
						slog.Debug("Reactor %s condition for path %s false", reactorItem.Reactor.ID(), eventQItem.Name)
						continue
					}
					eventsReactorCondition = append(eventsReactorCondition, eventQItem)
				}
				logger.Debug(fmt.Sprintf("Try reactor %s actions", reactorItem.Reactor.ID()))
				actionsResults, err := executePipelineActions(eventsReactorCondition, cfg.FileType, cfg.Path, reactorItem.ReactorActionsCfg, logger, reactorItem.Reactor.ID())
				if err != nil {
					logger.Error(err.Error())
				}
				for resultName, resultOut := range actionsResults {
					logger.Debug(fmt.Sprintf("(Reactor)%s:(Action)%s:%s", reactorItem.Reactor.ID(), resultName, resultOut))
				}
				eventsReactorCondition = eventsReactorCondition[:0]
				logger.Debug(fmt.Sprintf("Reset len of eventsConditions for next reactors. Len is %d", len(eventsReactorCondition)))
			}
			eventsQueue = eventsQueue[:0]
			logger.Debug(fmt.Sprintf("Reset events queue len. Len is %d", len(eventsQueue)))
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

func executePipelineActions(events []fsnotify.Event, cfgFileType string, cfgPath string, actionsCfg []config.Action, logger *slog.Logger, reactorName string) (map[string]string, error) {
	var actionExtArs = map[string]any{
		"config-file-type": cfgFileType,
		"config-path":      cfgPath,
		"reactorName":      reactorName,
		"events":           events,
	}
	var results = make(map[string]string)
	for _, actCfg := range actionsCfg {
		logger.Debug(fmt.Sprintf("Search action %s in coollection", actCfg.Name))
		act, err := action.NewAction(actCfg.Name)
		if err != nil {
			return nil, err
		}
		actArgs, err := act.PrepareArgs(logger, actionExtArs, actCfg.Args)
		if err != nil {
			return nil, err
		}
		out, err := act.Execute(*actArgs, logger)
		if err != nil {
			return nil, err
		}
		results[actCfg.Name] = out
	}
	return results, nil
}
