package action

import (
	"fmt"

	"github.com/vberlabs/inotify-watchman/internal/action/tarutil"
)

func NewAction(name string) (Action, error) {
	switch name {
	case "hello-action":
		return NewHelloAction(), nil
	case "archive-simple":
		return tarutil.NewArchiveSimple(), nil
	default:
		return nil, fmt.Errorf("Unknown action %s", name)
	}
}
