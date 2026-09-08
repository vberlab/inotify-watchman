package action

import (
	"fmt"
)

func NewAction(name string) (Action, error) {
	switch name {
	case "hello-action":
		return NewHelloAction(), nil
	default:
		return nil, fmt.Errorf("Unknown action %s", name)
	}
}
