package conf

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/xslasd/goxf/utils/xcast"
	"github.com/xslasd/goxf/utils/xmap"
	"golang.org/x/term"
)

var NoConfigErr = errors.New("no config...")
var InvalidConfigSource = errors.New("invalid config data source")
var InvalidConfigData = errors.New("invalid config data,encrypt or decrypt failed")
var UnmarshalInvalid = errors.New("unmarshal `func([]byte, any) error` invalid")

type Conf struct {
	sync.RWMutex
	keyDelimiter string
	keysMap      map[string]any
	watchers     map[string]func(*Conf)
}

type Unmarshal func([]byte, any) error

func NewConf(opts ...Option) *Conf {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}
	delimiter := opt.keyDelimiter
	if delimiter == "" {
		delimiter = "."
	}
	return &Conf{
		keysMap:      make(map[string]any),
		watchers:     make(map[string]func(*Conf)),
		keyDelimiter: delimiter,
	}
}

// SetKeyDelimiter set keyDelimiter of a defaultConfiguration instance.
func (c *Conf) SetKeyDelimiter(delimiter string) {
	c.keyDelimiter = delimiter
}

func (c *Conf) LoadFromConfigSource(ds ConfigSource, unmarshal Unmarshal) error {
	content, dsFormat, err := ds.ReadConfig()
	if err != nil {
		return err
	}

	if unmarshal == nil && dsFormat != "" {
		unmarshal = ExtToUnmarshal(dsFormat)
	}
	if unmarshal == nil {
		return UnmarshalInvalid
	}

	err = c.load(content, unmarshal)
	if err != nil {
		return err
	}

	// 仅在数据源开启了监听（Changed Channel 非空）时才启动后台热重载协程，防止 nil channel 导致协程永久泄漏
	if ch := ds.Changed(); ch != nil {
		go func() {
			for range ch {
				if content, dsFmt, err := ds.ReadConfig(); err == nil {
					um := unmarshal
					if um == nil && dsFmt != "" {
						um = ExtToUnmarshal(dsFmt)
					}
					if um != nil {
						_ = c.load(content, um)
					}
				}
			}
		}()
	}
	return nil
}

func (c *Conf) load(content []byte, unmarshal Unmarshal) error {
	configuration := make(map[string]any)
	err := unmarshal(content, &configuration)
	if err != nil {
		return err
	}
	return c.apply(configuration)
}

func (c *Conf) apply(conf map[string]any) error {
	c.Lock()
	defer c.Unlock()
	var changes = make(map[string]any)
	xmap.MergeStringMapWithChanged(c.keysMap, conf, changes, "")
	if len(changes) > 0 {
		c.notifyChanges(changes)
	}
	return nil
}

// SetWatcher set a func in a watcher with name of key
func (c *Conf) SetWatcher(key string, handle func(*Conf)) {
	c.watchers[key] = handle
}

func (c *Conf) notifyChanges(changes map[string]any) {
	var changedWatchPrefixMap = map[string]struct{}{}
	for watchPrefix := range c.watchers {
		for key := range changes {
			if strings.HasPrefix(key, watchPrefix) {
				changedWatchPrefixMap[watchPrefix] = struct{}{}
			}
		}
	}
	for changedWatchPrefix := range changedWatchPrefixMap {
		go c.watchers[changedWatchPrefix](c)
	}
}

// a.b.c
func (c *Conf) GetString(key string) string {
	return xcast.ToString(c.get(key))
}

func (c *Conf) GetBool(key string) bool {
	return xcast.ToBool(c.get(key))
}

func (c *Conf) GetInt(key string) int {
	return xcast.ToInt(c.get(key))
}

func (c *Conf) GetInt64(key string) int64 {
	return xcast.ToInt64(c.get(key))
}

func (c *Conf) GetFloat64(key string) float64 {
	return xcast.ToFloat64(c.get(key))
}

func (c *Conf) GetTime(key string) time.Time {
	return xcast.ToTime(c.get(key))
}

func (c *Conf) GetDuration(key string) time.Duration {
	return xcast.ToDuration(c.get(key))
}

func (c *Conf) GetStringSlice(key string) []string {
	return xcast.ToStringSlice(c.get(key))
}

func (c *Conf) GetSlice(key string) []any {
	return xcast.ToSlice(c.get(key))
}

func (c *Conf) GetStringMap(key string) map[string]any {
	return xcast.ToStringMap(c.get(key))
}

func (c *Conf) GetStringMapString(key string) map[string]string {
	return xcast.ToStringMapString(c.get(key))
}

func (c *Conf) GetSliceStringMap(key string) []map[string]any {
	return xcast.ToSliceStringMap(c.get(key))
}

func (c *Conf) GetStringMapStringSlice(key string) map[string][]string {
	return xcast.ToStringMapStringSlice(c.get(key))
}

func (c *Conf) get(key string) any {
	paths := strings.Split(key, c.keyDelimiter)
	c.RLock()
	defer c.RUnlock()
	val := traverse(c.keysMap, paths, 0)
	return val
}

func traverse(items map[string]any, paths []string, index int64) any {
	item, ok := items[paths[index]]
	if ok {
		switch item.(type) {
		case map[string]any:
			index++
			if index > int64(len(paths))-1 {
				return item
			}
			k := item.(map[string]any)
			return traverse(k, paths, index)
		default:
			return item
		}
	}
	return nil
}

// ErrInvalidKey ...
var ErrInvalidKey = errors.New("invalid key, maybe not exist in config")

func (c *Conf) UnmarshalKey(key string, rawVal any) error {
	value := c.get(key)
	if value == nil {
		return errors.Wrap(ErrInvalidKey, key)
	}
	return mapstructure.Decode(value, rawVal)
}

func verifyPassword(configAddr string, pwd string, cryptFn func(string) string, isEnc bool) bool {
	if pwd != "" {
		vPassword := false
		for _, arg := range os.Args {
			if arg == "--crypt-conf" || arg == "-e" {
				vPassword = true
				break
			}
		}
		if vPassword {
			fmt.Println()
			fmt.Printf("Enter config password for [%s]: ", configAddr)
			var inputPwd string
			if term.IsTerminal(int(os.Stdin.Fd())) {
				pwdBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					fmt.Printf("Failed to read password for [%s]: %v\n", configAddr, err)
					os.Exit(1)
				}
				inputPwd = string(pwdBytes)
			} else {
				fmt.Scanln(&inputPwd)
			}

			if inputPwd == "" {
				fmt.Printf("Password for [%s] cannot be empty\n", configAddr)
				os.Exit(1)
			}
			checkPwd := inputPwd
			if cryptFn != nil {
				checkPwd = cryptFn(inputPwd)
			}
			if checkPwd != pwd {
				fmt.Printf("Password for [%s] is not correct\n", configAddr)
				os.Exit(1)
			}
			return true
		}
		if !isEnc {
			fmt.Printf("invalid config data decrypt failed for [%s]\n", configAddr)
			os.Exit(1)
		}
	}
	return false
}
