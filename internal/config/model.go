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
	Reactor        []Reactor `yaml:"reactor"`
	EventQueueSize int       `yaml:"event_queue_size"`
}

type Reactor struct {
	Name             string         `yaml:"name"`
	Args             map[string]any `yaml:"args"`
	Actions          []Action       `yaml:"actions"`
	IgnoreErrors     bool           `yaml:"ignore_errors"`
	IgnoreArgsErrors bool           `yaml:"ignore_args_errors"`
}

type Action struct {
	Name             string         `yaml:"name"`
	Args             map[string]any `yaml:"args"`
	IgnoreErrors     bool           `yaml:"ignore_errors"`
	IgnoreArgsErrors bool           `yaml:"ignore_args_errors"`
}
