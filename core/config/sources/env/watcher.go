package env

import (
	"context"

	"github.com/kweaver-ai/idrm-go-frame/core/config/sources"
)

type watcher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var _ sources.Watcher = (*watcher)(nil)

func NewWatcher() (sources.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{ctx: ctx, cancel: cancel}, nil
}

// Next will be blocked until the Stop method is called
func (w *watcher) Next() ([]*sources.KeyValue, error) {
	<-w.ctx.Done()
	return nil, w.ctx.Err()
}

func (w *watcher) Stop() error {
	w.cancel()
	return nil
}
