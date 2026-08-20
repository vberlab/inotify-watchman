package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Events fsnotify.Op

var EventsMap = map[string]fsnotify.Op{
	"create": fsnotify.Create,
	"modify": fsnotify.Write,
	"remove": fsnotify.Remove,
	"rename": fsnotify.Rename,
	"chmod":  fsnotify.Chmod,
}

func (e *Events) UnmarshalYAML(node *yaml.Node) error {
	var event fsnotify.Op
	var hasKey bool

	if (node.Kind != yaml.ScalarNode) || (node.Tag != "!!str") {
		return fmt.Errorf("Event type not a string value")
	}
	event, hasKey = EventsMap[node.Value]

	if !hasKey {
		return fmt.Errorf("Unknown event type: %s", node.Value)
	}

	*e = Events(event)
	return nil
}
