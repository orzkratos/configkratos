package configkratos

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/yyle88/erero"
	"github.com/yyle88/must"
)

type DataSource struct {
	data    []byte
	format  string
	watcher *ConfigWatcher
}

// NewDataSource create a source with byte slice
// param format range: https://github.com/go-kratos/kratos/blob/main/encoding/encoding.go#L21
// param format range: https://github.com/go-kratos/kratos/blob/c82f7957223f7e7744fec01f8466dea7bf6ae6fd/encoding/encoding.go#L25
// current range is in: "json" "yaml" "form" "xml" "proto"
func NewDataSource(data []byte, format string) *DataSource {
	return &DataSource{
		data:    data,
		format:  must.Nice(format),
		watcher: nil, //不在 new 函数里初始化它，因为没有close的地方，就不要初始化避免泄漏东西
	}
}

func NewYamlSource(data []byte) *DataSource {
	return NewDataSource(data, "yaml")
}

func NewJsonSource(data []byte) *DataSource {
	return NewDataSource(data, "json")
}

func (p *DataSource) Load() ([]*config.KeyValue, error) {
	kv := &config.KeyValue{
		Key:    "config",
		Value:  p.data,
		Format: p.format,
	}
	return []*config.KeyValue{kv}, nil
}

func (p *DataSource) Watch() (config.Watcher, error) {
	if p.watcher != nil {
		return nil, erero.New("REPEAT-WATCH") //正常情况下只会watch一次，当同一个 source 两次传给 config.WithSource() 时，就会有这个问题
	}
	watcher, err := NewConfigWatcher(p.format)
	if err != nil {
		return nil, erero.Wro(err)
	}
	p.watcher = watcher
	return watcher, nil
}

func (p *DataSource) Update(data []byte) error {
	if p.watcher == nil {
		return erero.New("NOT-WATCHING")
	}
	if err := p.watcher.update(data); err != nil {
		return erero.Wro(err)
	}
	return nil
}

type ConfigWatcher struct {
	dataChan   chan []byte
	format     string
	ctx        context.Context
	cancelFunc context.CancelFunc
	mutex      *sync.Mutex
	isWatching bool
}

func NewConfigWatcher(format string) (*ConfigWatcher, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	res := &ConfigWatcher{
		dataChan:   make(chan []byte, 1), // 缓冲通道，避免阻塞
		format:     must.Nice(format),
		ctx:        ctx,
		cancelFunc: cancelFunc,
		mutex:      &sync.Mutex{},
		isWatching: true,
	}
	return res, nil
}

func (w *ConfigWatcher) update(data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isWatching {
		return erero.New("WATCHER-IS-STOPED. CAN-NOT-UPDATE-DATA")
	}

	w.dataChan <- data
	return nil
}

func (w *ConfigWatcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err() // 返回 context.Canceled，在外层等的就是这个返回
	case data, ok := <-w.dataChan:
		if !ok {
			return nil, context.Canceled // dataChan 关闭时返回 context.Canceled，因为外面等的就是这个返回
		}
		oneItem := &config.KeyValue{
			Key:    "config", // 这里感觉是这样就行，但具体也不明白，还能填什么别的
			Value:  data,
			Format: w.format,
		}
		return []*config.KeyValue{oneItem}, nil
	}
}

func (w *ConfigWatcher) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isWatching {
		return erero.New("WATCHER-IS-STOPED. CAN-NOT-STOP-AGAIN")
	}
	w.isWatching = false

	w.cancelFunc()    // 触发 ctx.Done()
	close(w.dataChan) // 关闭 dataChan，确保资源清理
	return nil
}
