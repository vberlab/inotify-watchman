package actions

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
)

type HelloAction struct {
	actions_core.ActionBase
}

func NewHelloAction() *HelloAction {
	var expectedArgs actions_core.ExpectedArgsMap = actions_core.ExpectedArgsMap{
		"hello-text": reflect.TypeOf(""),
		"files":      reflect.TypeOf([]string(nil)),
	}
	return &HelloAction{
		ActionBase: actions_core.NewActionBase("hello-action", expectedArgs, []string{}),
	}
}

func (h *HelloAction) RunAction(args actions_core.ActionArgs, logger *slog.Logger) ([]string, error) {
	var files []string
	logger.Info(fmt.Sprintf("Run Hello action. Hello text: %s", args.Args["hello-text"]))
	logger.Info("With files")
	for _, path := range args.Args["files"].([]string) {
		logger.Info(fmt.Sprintf("path: %s", path))
		files = append(files, path)
	}
	return files, nil
}
