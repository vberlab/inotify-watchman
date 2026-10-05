package reactor_registry

import (
	"fmt"

	"github.com/vberlabs/inotify-watchman/internal/reactor/reactor_core"
	"github.com/vberlabs/inotify-watchman/internal/reactor/reactors"
)

func NewReactor(name string) (reactor_core.Reactor, error) {
	switch name {
	case "hash-changed":
		return reactors.NewHashChanged(), nil
	default:
		return nil, fmt.Errorf("Unknown reactor %s", name)
	}
}
