package reactor

import "github.com/vberlabs/inotify-watchman/internal/reactor"

var ReactorCollection = map[string]Reactor{
	"hash-changed": reactor.HashChanged,
}
