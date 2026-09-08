package action

import "log/slog"

type ActionArgs struct {
	Args map[string]any
}

func NewActionArgs() *ActionArgs {
	return &ActionArgs{
		Args: make(map[string]any),
	}
}

type Action interface {
	ID() string
	IgnoreArgsErrors() bool
	ExpectedArgs() []string
	PrepareArgs(logger *slog.Logger, args ...map[string]any) (*ActionArgs, error)
	Execute(args ActionArgs, logger *slog.Logger) (string, error)
}
