package internal

import (
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/vberlabs/inotify-watchman/internal/config"
)

func WatcherHead(config config.Tracking, exitOnError bool) error {
	// Init fsnotify.Wather
	var watcher *fsnotify.Watcher
	var watcherError error
	watcher, watcherError = fsnotify.NewWatcher()
	if watcherError != nil {
		return watcherError
	}
	defer watcher.Close()

	watcherError = watcher.Add(config.Path)
	if watcherError != nil {
		return watcherError
	}

	// Watch events
	var eventRecived bool
	var event fsnotify.Event
	for {
		eventRecived, watcherError = watchEvents(watcher, &event, config.Events)
		if watcherError != nil {
			if exitOnError {
				return watcherError
			}
		}
		if eventRecived {
			execReactor(&event, config)
		}
	}
}

func watchEvents(watcher *fsnotify.Watcher, event *fsnotify.Event, events []config.Events) (bool, error) {
	for {
		select {
		case recivedEvent := <-watcher.Events:
			slog.Debug("Event recived", "event", event)
			for _, eventType := range events {
				if recivedEvent.Has(fsnotify.Op(eventType)) {
					*event = recivedEvent
					return true, nil
				}
			}
		case err := <-watcher.Errors:
			slog.Error("Watcher recived error", "error", err)
			return false, err
		}
	}
}

func execReactor(event *fsnotify.Event, config config.Tracking) {
	fmt.Printf("Execute reactor on %s\n", event.Name)

}
