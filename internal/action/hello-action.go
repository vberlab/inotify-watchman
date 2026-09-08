package action

import (
	"errors"
	"fmt"
	"log/slog"
)

type HelloAction struct {
	actionID         string
	helloText        string
	ignoreArgsErrors bool
}

func NewHelloAction() *HelloAction {
	return &HelloAction{
		actionID:         "hellow-action",
		helloText:        "Hello, world!",
		ignoreArgsErrors: true,
	}
}

func (a *HelloAction) ID() string {
	return a.actionID
}

func (a *HelloAction) IgnoreArgsErrors() bool {
	return a.ignoreArgsErrors
}

func (a *HelloAction) ExpectedArgs() []string {
	return nil
}

func (a *HelloAction) PrepareArgs(logger *slog.Logger, args ...map[string]any) (*ActionArgs, error) {
	logger.Debug("Hello action PrepareArgs")
	var actionArgs = NewActionArgs()
	if !a.ignoreArgsErrors {
		if len(args) == 0 {
			return nil, errors.New("Arguments not passed")
		}
		if len(a.ExpectedArgs()) == 0 {
			return nil, errors.New("Expected arguments not specified")
		}
	}
	for _, item := range args {
		logger.Debug("Place hello action args to ActionArgs struct")
		for k, v := range item {
			logger.Debug(fmt.Sprintf("Place args %s", k))
			actionArgs.Args[k] = v
		}
	}
	return actionArgs, nil
}

func (a *HelloAction) Execute(args ActionArgs, logger *slog.Logger) (string, error) {
	logger.Debug("Hello action execute")
	return a.helloText, nil
}
