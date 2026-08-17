package config

type Config struct {
	Tracking []Tracking
}

type Tracking struct {
	Path     string
	FileType string
	Events   []Events
	Pipeline Pipeline
}

type Pipeline struct {
	Reactor []Reactor
}

type Reactor struct {
	Name    string
	Args    map[string]any
	Actions []Action
}

type Action struct {
	Name string
	Args map[string]any
}
