package config

import (
	"github.com/kweaver-ai/idrm-go-frame/core/config/sources"
	"github.com/kweaver-ai/idrm-go-frame/core/config/sources/env"
	"github.com/kweaver-ai/idrm-go-frame/core/config/sources/file"
)

var manager *Manager

// Manager config manager, hold the config
type Manager struct {
	Config
}

// InitSources init config source, with env
func InitSources(paths ...string) {
	sources := make([]sources.Source, len(paths)+1)
	sources[0] = env.NewSource()
	for i, path := range paths {
		sources[i+1] = file.NewSource(path)
	}
	Init(sources...)
}

// Init config source, inconvenient for user, because of circular reference
func Init(sources ...sources.Source) {
	c := New(WithSource(sources...))
	if err := c.Load(); err != nil {
		panic(err)
	}
	manager = &Manager{Config: c}
}

func Load() {
	if err := manager.Config.Load(); err != nil {
		panic(err)
	}
}

// Scan any type
func Scan[T any](keys ...string) T {
	key := ""
	if len(keys) > 0 && keys[0] != "" {
		key = keys[0]
	}
	var data T
	if key == "" {
		if err := manager.Config.Scan(&data); err != nil {
			panic(err)
		}
		return data
	}
	value := manager.Config.Value(key)
	if err := value.Scan(&data); err != nil {
		panic(err)
	}
	return data
}

func GetValue(key string) Value {
	return manager.Config.Value(key)
}

func Watch(key string, o Observer) error {
	return manager.Config.Watch(key, o)
}

func Close() error {
	return manager.Config.Close()
}
