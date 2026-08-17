package config

import "github/fsnotify/fsnotify"

var EventsMap = map[string]fsnotify.Op{
	"create": fsnotify.Create,
	"modify": fsnotify.Write,
	"remove": fsnotify.Remove,
	"rename": fsnotify.Rename,
	"chmod": fsnotify.chmod
}