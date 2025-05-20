package configkratos

import (
	"context"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/yyle88/erero"
	"github.com/yyle88/must"
)

// DataStatic is a static data source that does not support dynamic updates. It is suitable to load configurations that remain static.
// DataStatic 因为发现那个带 Watch 的，没有监听全局任意变化的操作，因此还得手动 Scan 数据，感觉不是很好用，就实现了个纯静态的，不再监听变动，让逻辑更加轻便
type DataStatic struct {
	data   []byte
	format string
}

func NewDataStatic(data []byte, format string) *DataStatic {
	return &DataStatic{
		data:   data,
		format: must.Nice(format),
	}
}

func NewYamlStatic(data []byte) *DataStatic {
	return NewDataStatic(data, "yaml")
}

func NewJsonStatic(data []byte) *DataStatic {
	return NewDataStatic(data, "json")
}

func (p *DataStatic) Load() ([]*config.KeyValue, error) {
	kv := &config.KeyValue{
		Key:    "config",
		Value:  p.data,
		Format: p.format,
	}
	return []*config.KeyValue{kv}, nil
}

func (p *DataStatic) Watch() (config.Watcher, error) {
	w, err := NewStaticWatcher()
	if err != nil {
		return nil, erero.Wro(err)
	}
	return w, nil
}

// StaticWatcher means no watch. cp from: https://github.com/go-kratos/kratos/blob/dbd7664eff951d1e13b291d95a226f1e9101c8e1/config/env/watcher.go#L11
type StaticWatcher struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewStaticWatcher() (config.Watcher, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &StaticWatcher{
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}, nil
}

func (w *StaticWatcher) Next() ([]*config.KeyValue, error) {
	<-w.ctx.Done()
	return nil, w.ctx.Err()
}

func (w *StaticWatcher) Stop() error {
	w.cancelFunc()
	return nil
}
