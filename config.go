package watcher

type Monitor struct {
	Directory string   `json:"directory"`
	Exclude   []string `json:"exclude,omitempty"`
	Action    []string `json:"action"`
}

type Config struct {
	Monitor []Monitor `json:"monitor"`
}
