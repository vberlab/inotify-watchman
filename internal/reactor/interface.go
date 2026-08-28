package reactor

type Reactor interface {
	ID() string
	CheckCondition(args map[string]any) (bool, error)
}
