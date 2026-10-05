package reactor_core

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"

	"github.com/vberlabs/inotify-watchman/internal/config"
)

// Base structures of Reactor interface
type ReactorBase struct {
	reactorID        string
	ignoreErrors     bool
	ignoreArgsErrors bool
	expectedArgs     ExpectedArgsMap
}

func NewReactorBase(name string, expectedArgs ExpectedArgsMap) ReactorBase {
	return ReactorBase{
		reactorID:    name,
		expectedArgs: expectedArgs,
	}
}

// Expected args map
type ExpectedArgsMap map[string]reflect.Type

// Complited arguments for CheckCondition function
type ReactorArgs struct {
	Args map[string]any
}

func NewReactorArgs() *ReactorArgs {
	return &ReactorArgs{
		Args: make(map[string]any),
	}
}

// Reactor interfaces context for multiple reactor pipeline
type ReactorContext struct {
	Reactor       Reactor
	ReactorConfig config.Reactor
}

type Reactor interface {
	ID() string
	ExpectedArgs() ExpectedArgsMap
	ExpectedArgsKeys() []string
	Configure(ignoreErrors bool, ignoreArgsErrors bool)
	PreapreArgs(logger *slog.Logger, args ...map[string]any) (*ReactorArgs, error)
	CheckCondition(logger *slog.Logger, args ReactorArgs) ([]string, error)
}

// Implement base methods
func (r *ReactorBase) ID() string {
	return r.reactorID
}

func (r *ReactorBase) IgnoreErrors() bool {
	return r.ignoreErrors
}

func (r *ReactorBase) IgnoreArgsErrors() bool {
	return r.ignoreArgsErrors
}

// List of keys with expected argument for reactor CheckCondition
func (r *ReactorBase) ExpectedArgs() ExpectedArgsMap {
	return r.expectedArgs
}

func (r *ReactorBase) ExpectedArgsKeys() []string {
	var keys []string
	for k, _ := range r.ExpectedArgs() {
		keys = append(keys, k)
	}
	return keys
}

// Function for errors processing. Set in configuration file
func (r *ReactorBase) Configure(ignoreErrors bool, ignoreArgsErrors bool) {
	r.ignoreErrors = ignoreErrors
	r.ignoreArgsErrors = ignoreArgsErrors
}

// Prepare CheckCondition arguments
func (r *ReactorBase) PreapreArgs(logger *slog.Logger, args ...map[string]any) (*ReactorArgs, error) {
	var (
		ErrorExpectedArgumentNotSpecified error = errors.New("Expected argument not specified")
		argFound                          bool  = false
		argValue                          any
		reactorArgs                       *ReactorArgs = NewReactorArgs()
		foundExpectedArgs                 []string
		actualType                        reflect.Type
	)
	// Find expected args
	logger.Debug(fmt.Sprintf("Prepare %s reactor arguments", r.ID()))
	for _, mapItem := range args {
		for expectedArgN, expectedArgT := range r.ExpectedArgs() {
			argValue, argFound = mapItem[expectedArgN]
			if !argFound {
				continue
			}
			actualType = reflect.TypeOf(argValue)
			if actualType != expectedArgT {
				return nil, fmt.Errorf("argument %q has type %s, expected %s", expectedArgN, actualType, expectedArgT)
			}
			reactorArgs.Args[expectedArgN] = argValue
			foundExpectedArgs = append(foundExpectedArgs, expectedArgN)
		}
	}
	// Check all expected args found
	logger.Debug("Check founded arguments")
	for _, expectedArg := range r.ExpectedArgsKeys() {
		if !slices.Contains(foundExpectedArgs, expectedArg) {
			return nil, fmt.Errorf("%w, %q", ErrorExpectedArgumentNotSpecified, expectedArg)
		}
	}
	return reactorArgs, nil
}
