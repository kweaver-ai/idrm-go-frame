package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/config/sources"
)

var _ sources.Source = (*file)(nil)

type file struct {
	path string
}

// NewSource new a file source.
func NewSource(path string) sources.Source {
	return &file{path: path}
}

func (f *file) loadFile(path string) (*sources.KeyValue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &sources.KeyValue{
		Key:    info.Name(),
		Format: format(info.Name()),
		Value:  data,
	}, nil
}

func (f *file) loadDir(path string) (kvs []*sources.KeyValue, err error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	envPrefix := AppendEnv(sources.DefaultPrefix) + "."
	for _, file := range files {
		fileName := file.Name()
		// ignore hidden files
		if file.IsDir() || strings.HasPrefix(fileName, ".") {
			continue
		}
		if strings.HasPrefix(fileName, sources.DefaultPrefix) && !strings.HasPrefix(fileName, envPrefix) {
			continue
		}
		kv, err := f.loadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}
	return
}

func (f *file) Load() (kvs []*sources.KeyValue, err error) {
	fi, err := os.Stat(f.path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return f.loadDir(f.path)
	}
	kv, err := f.loadFile(f.path)
	if err != nil {
		return nil, err
	}
	return []*sources.KeyValue{kv}, nil
}

func (f *file) Watch() (sources.Watcher, error) {
	return newWatcher(f)
}

func projectEnv() string {
	env, has := os.LookupEnv(sources.ProjectEnvKey)
	if has {
		return strings.ToLower(env)
	}
	return ""
}

// AppendEnv  根据项目的环境，给出不同的
func AppendEnv(path string) string {
	env := projectEnv()
	if env == "" {
		return path
	}
	pos := strings.LastIndex(path, ".")
	if pos < 0 {
		return fmt.Sprintf("%s_%s", path, env)
	}
	return fmt.Sprintf("%s_%s%s", path[:pos], env, path[pos:])
}
