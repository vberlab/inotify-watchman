package config

type Config struct {
	WatcherExitOnError bool       `yaml:"watcher_exit_on_error"`
	Tracking           []Tracking `yaml:"tracking"`
}

type Tracking struct {
	Path           string   `yaml:"path"`
	FileType       string   `yaml:"type"`
	FileNameFilter string   `yaml:"file_name_filter"`
	Events         []Events `yaml:"events"`
	Pipeline       Pipeline `yaml:"pipeline"`
}

type Pipeline struct {
	Reactor []Reactor `yaml:"reactor"`
}

type Reactor struct {
	Name    string         `yaml:"name"`
	Args    map[string]any `yaml:"args"`
	Actions []Action       `yaml:"actions"`
}

type Action struct {
	Name string         `yaml:"name"`
	Args map[string]any `yaml:"args"`
}
