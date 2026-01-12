package env

import (
	"os"
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/config/sources"
)

type env struct {
	prefixs []string
}

func NewSource(prefixSlice ...string) sources.Source {
	prefix := os.Getenv(sources.ProjectPrefix)
	if prefix != "" && len(prefixSlice) <= 0 {
		prefixSlice = strings.Split(prefix, ";")
	}
	return &env{prefixs: prefixSlice}
}

func (e *env) Load() (kv []*sources.KeyValue, err error) {
	return e.load(os.Environ()), nil
}

func (e *env) load(envStrings []string) []*sources.KeyValue {
	var kv []*sources.KeyValue
	for _, envStr := range envStrings {
		var k, v string
		subs := strings.SplitN(envStr, "=", 2) //nolint:gomnd
		k = subs[0]
		if len(subs) > 1 {
			v = subs[1]
		}

		if len(e.prefixs) > 0 {
			p, ok := matchPrefix(e.prefixs, k)
			if !ok || len(p) == len(k) {
				continue
			}
			// trim prefix
			k = strings.TrimPrefix(k, p)
			k = strings.TrimPrefix(k, "_")
		}

		if len(k) != 0 {
			kv = append(kv, &sources.KeyValue{
				Key:   k,
				Value: []byte(v),
			})
		}
	}
	return kv
}

func (e *env) Watch() (sources.Watcher, error) {
	w, err := NewWatcher()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func matchPrefix(prefixes []string, s string) (string, bool) {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return p, true
		}
	}
	return "", false
}
