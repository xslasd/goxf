package filesource

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/utils/xfmt"
)

type ConfigSource struct {
	path        string
	changed     chan struct{}
	done        chan struct{}
	delay       chan struct{}
	isWatch     bool
	notifyDelay time.Duration
	closeOnce   sync.Once
}

func NewConfigSource(path string, watch bool) *ConfigSource {
	ds := &ConfigSource{
		path:        path,
		notifyDelay: 5 * time.Second,
		isWatch:     watch,
	}

	if ds.isWatch {
		ds.changed = make(chan struct{}, 1)
		ds.done = make(chan struct{})
		ds.delay = make(chan struct{}, 1)
		go ds.watch()
	}
	return ds
}

func (s *ConfigSource) SetDelay(d time.Duration) {
	s.notifyDelay = d
}

func (s *ConfigSource) ReadConfig() ([]byte, string, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		return nil, "", err
	}
	// 支持环境变量扩展
	return []byte(os.ExpandEnv(string(content))), filepath.Ext(s.path), nil
}

func (s *ConfigSource) Changed() <-chan struct{} {
	return s.changed
}

func (s *ConfigSource) Close() error {
	s.closeOnce.Do(func() {
		if s.isWatch {
			close(s.done)
		}
	})
	return nil
}

func (s *ConfigSource) watch() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	// 监听文件所在目录，以应对编辑器 atomic save 导致 inode 变更的问题
	dir := filepath.Dir(s.path)
	err = w.Add(dir)
	if err != nil {
		xfmt.Printf("file watch failed: %s", err.Error())
		return
	}

	go func() {
		defer w.Close()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				// 监听写入、创建、重命名操作（重命名在某些编辑器的保存机制中很常见）
				const mask = fsnotify.Write | fsnotify.Create | fsnotify.Rename
				if event.Op&mask != 0 && filepath.Clean(event.Name) == filepath.Clean(s.path) {
					xfmt.Printf("modified file %s, event: %s", event.Name, event.Op)
					select {
					case s.delay <- struct{}{}:
					default:
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				xfmt.Printf("file watch error: %s", err.Error())
			case <-s.done:
				return
			}
		}
	}()

	go func() {
		var t *time.Timer
		for {
			select {
			case <-s.delay:
				if t != nil {
					t.Stop()
				}
				// 延迟触发通知，防抖处理
				t = time.AfterFunc(s.notifyDelay, func() {
					select {
					case s.changed <- struct{}{}:
					case <-s.done:
					default:
					}
				})
			case <-s.done:
				if t != nil {
					t.Stop()
				}
				// 停止后关闭通知通道，通知监听者退出
				close(s.changed)
				return
			}
		}
	}()
}

func init() {
	conf.Register(conf.FileScheme, func(configAddr string, isWatch bool) conf.ConfigSource {
		return NewConfigSource(configAddr, isWatch)
	})
}
