package actions_core

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"

	internal_misc "github.com/vberlabs/inotify-watchman/internal/misc"
)

type ExpectedArgsMap map[string]reflect.Type
type ActionArgs struct {
	Args map[string]any
}

func NewActionArgs() *ActionArgs {
	return &ActionArgs{
		Args: make(map[string]any),
	}
}

type ActionBase struct {
	actionID         string
	ignoreErrors     bool
	ignoreArgsErrors bool
	expectedArgs     ExpectedArgsMap
	requireArgs      []string
}

func NewActionBase(name string, expectedArgs ExpectedArgsMap, requiredArgs []string) ActionBase {
	return ActionBase{
		actionID:     name,
		expectedArgs: expectedArgs,
		requireArgs:  requiredArgs,
	}
}

type Action interface {
	ID() string
	IgnoreErrors() bool
	IgnoreArgsErrors() bool
	Configure(ignoreErrors bool, ignoreArgsErrors bool)
	ExpectedArgs() ExpectedArgsMap
	ExpectedArgsKeys() []string
	RequiredArgs() []string
	PreapreArgs(logger *slog.Logger, args ...map[string]any) (*ActionArgs, error)
	RunAction(args ActionArgs, logger *slog.Logger) ([]string, error)
}

func (a *ActionBase) ID() string {
	return a.actionID
}

func (a *ActionBase) IgnoreErrors() bool {
	return a.ignoreErrors
}

func (a *ActionBase) IgnoreArgsErrors() bool {
	return a.ignoreArgsErrors
}

func (a *ActionBase) Configure(ignoreErrors bool, ignoreArgsErrors bool) {
	a.ignoreErrors = ignoreErrors
	a.ignoreArgsErrors = ignoreArgsErrors
}

func (a *ActionBase) ExpectedArgs() ExpectedArgsMap {
	return a.expectedArgs
}

func (r *ActionBase) ExpectedArgsKeys() []string {
	var keys []string
	for k, _ := range r.ExpectedArgs() {
		keys = append(keys, k)
	}
	return keys
}

func (a *ActionBase) RequiredArgs() []string {
	return a.requireArgs
}

func (a *ActionBase) PreapreArgs(logger *slog.Logger, args ...map[string]any) (*ActionArgs, error) {
	var (
		ErrorExpectedArgumentNotSpecified error = errors.New("Expected argument not specified")
		argFound                          bool  = false
		argValue                          any
		actionArgs                        *ActionArgs = NewActionArgs()
		foundExpectedArgs                 []string
		actualType                        reflect.Type
		err                               error
	)
	// Find expected args
	logger.Debug(fmt.Sprintf("Prepare %s action arguments", a.ID()))
	for _, mapItem := range args {
		for expectedArgN, expectedArgT := range a.ExpectedArgs() {
			argValue, argFound = mapItem[expectedArgN]
			if !argFound {
				continue
			}
			argValue, err = internal_misc.ConvertArgument(argValue, expectedArgT)
			if err != nil {
				return nil, fmt.Errorf(
					"argggument %q: %w",
					expectedArgN,
					expectedArgT,
				)
			}
			actualType = reflect.TypeOf(argValue)
			if actualType != expectedArgT {
				return nil, fmt.Errorf("argument %q has type %s, expected %s", expectedArgN, actualType, expectedArgT)
			}
			actionArgs.Args[expectedArgN] = argValue
			foundExpectedArgs = append(foundExpectedArgs, expectedArgN)
		}
	}
	// Check all expected args found
	logger.Debug("Check founded arguments")
	for _, expectedArg := range a.ExpectedArgsKeys() {
		if !slices.Contains(foundExpectedArgs, expectedArg) {
			if !slices.Contains(a.RequiredArgs(), expectedArg) {
				continue
			}
			return nil, fmt.Errorf("%w, %q", ErrorExpectedArgumentNotSpecified, expectedArg)
		}
	}
	return actionArgs, nil
}
