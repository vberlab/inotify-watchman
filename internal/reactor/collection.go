package reactor

import "fmt"

func NewReactor(name string) (Reactor, error) {
	switch name {
	case "hash-changed":
		return NewHashChanged(), nil
	default:
		return nil, fmt.Errorf("Unknown reactor: %s", name)
	}
}
