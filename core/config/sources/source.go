package sources

const (
	//ProjectEnvKey recognize env of project, dev, release, production
	ProjectEnvKey = "PROJECT_ENV"
	//ProjectPrefix project env key prefix
	ProjectPrefix = "PROJECT_PREFIX"
	//DefaultPrefix  default config file name prefix
	DefaultPrefix = "config"
)

// KeyValue is config key value.
type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}

// Source is config source.
type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}
