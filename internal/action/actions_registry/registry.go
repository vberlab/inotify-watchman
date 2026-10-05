package actions_registry

import (
	"fmt"

	"github.com/vberlabs/inotify-watchman/internal/action/actions"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
)

func NewAction(name string) (actions_core.Action, error) {
	switch name {
	case "hello-action":
		return actions.NewHelloAction(), nil
	case "archive-file":
		return actions.NewArchiveFileAction(), nil
	case "write-vault":
		return actions.NewWriteVaultAction(), nil
	default:
		return nil, fmt.Errorf("Unknown action %s", name)
	}
}
