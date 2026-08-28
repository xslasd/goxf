package conf

import (
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/utils/xfile"
)

const FileScheme = "file"

type ConfigSource interface {
	ReadConfig() ([]byte, string, error) //读取配置内容及格式(如 "yaml", ".json")
	Changed() <-chan struct{}            //监听配置是否改动
	io.Closer                            //io 关闭
}

var (
	registry = make(map[string]DataSourceCreatorFunc)
)

// DataSourceCreatorFunc represents a dataSource creator function
type DataSourceCreatorFunc func(string, bool) ConfigSource

// Register registers a dataSource creator function to the registry
func Register(scheme string, creator DataSourceCreatorFunc) {
	registry[scheme] = creator
}

// CreateConfigSource 根据配置地址创建并初始化对应的 ConfigSource 以及对应的 Unmarshal 函数
func CreateConfigSource(configAddr string, opts ...Option) (ConfigSource, Unmarshal, error) {
	if configAddr == "" {
		return nil, nil, NoConfigErr
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	var scheme = FileScheme
	var path = configAddr

	if strings.Contains(configAddr, "://") {
		urlObj, err := url.Parse(configAddr)
		if err != nil {
			return nil, nil, err
		}
		if len(urlObj.Scheme) > 1 {
			scheme = urlObj.Scheme
		}
		path = urlObj.Path
		// Windows 下 /D:/path -> D:/path
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
	}

	creatorFunc, exist := GetDataSourceCreatorFunc(scheme)
	if !exist {
		return nil, nil, InvalidConfigSource
	}

	var ds ConfigSource
	var finalUnmarshal = opt.unmarshal
	watch := opt.watch
	effPwd, effCryptFn := opt.getEffectivePassword()

	switch scheme {
	case FileScheme:
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, err
		}

		baseName := filepath.Base(absPath)
		ext := filepath.Ext(baseName)
		dir := filepath.Dir(absPath)
		isEnc := strings.HasSuffix(baseName, ".enc")
		isLocal := strings.HasSuffix(baseName, ".local"+ext)

		if isEnc {
			watch = false
		}

		displayAddr := formatRelPath(absPath)
		vPwd := verifyPassword(displayAddr, effPwd, effCryptFn, isEnc)

		wrapper := &fileSourceWrapper{
			isEnc:           isEnc,
			isLocal:         isLocal,
			vPwd:            vPwd,
			ext:             ext,
			absPath:         absPath,
			dir:             dir,
			password:        effPwd,
			passwordCryptFn: effCryptFn,
			customUnmarshal: opt.unmarshal,
			changed:         make(chan struct{}, 1),
		}

		mainDs := creatorFunc(absPath, watch)
		wrapper.mainDs = mainDs

		if !isEnc && !isLocal {
			localPath := strings.TrimSuffix(absPath, ext) + ".local" + ext
			wrapper.localPath = localPath
			exists, _ := xfile.Exists(localPath)
			if exists {
				wrapper.localDs = creatorFunc(localPath, watch)
			}
		}

		wrapper.startWatch()

		ds = wrapper
		finalUnmarshal = nil // FileScheme wrapper determines the final unmarshal format (e.g. after yaml merge)

	default:
		ds = creatorFunc(configAddr, watch)
	}

	if watch {
		hooks.Register(hooks.Stage_AfterStop, func() {
			err := ds.Close()
			if err != nil {
				log.Errorf("close config source error: %v", err)
			}
		})
	}

	return ds, finalUnmarshal, nil
}

// NewConfFromSource 创建并返回一个全新的独立 Conf 实例，并从指定的配置源加载配置（不影响全局 defaultConfiguration）
func NewConfFromSource(configAddr string, opts ...Option) (*Conf, error) {
	ds, um, err := CreateConfigSource(configAddr, opts...)
	if err != nil {
		return nil, err
	}
	c := NewConf(opts...)
	if err := c.LoadFromConfigSource(ds, um); err != nil {
		return nil, err
	}
	return c, nil
}

// NewFileConfSource 从指定文件路径创建并返回一个全新的独立 Conf 实例（自动按文件后缀匹配解析器）
func NewFileConfSource(configAddr string, watch bool, opts ...Option) (*Conf, error) {
	allOpts := append([]Option{WithWatch(watch)}, opts...)
	return NewConfFromSource(configAddr, allOpts...)
}

// LoadFromSource 将指定地址的配置源加载并合并到当前 Conf 实例中
func (c *Conf) LoadFromSource(configAddr string, opts ...Option) error {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}
	if opt.keyDelimiter != "." && c.keyDelimiter == "." {
		c.keyDelimiter = opt.keyDelimiter
	}
	ds, um, err := CreateConfigSource(configAddr, opts...)
	if err != nil {
		return err
	}
	return c.LoadFromConfigSource(ds, um)
}

// LoadFromSource 加载指定地址的配置源并合并到全局默认 defaultConfiguration 实例中
func LoadFromSource(configAddr string, opts ...Option) error {
	return defaultConfiguration.LoadFromSource(configAddr, opts...)
}

func GetDataSourceCreatorFunc(scheme string) (DataSourceCreatorFunc, bool) {
	source, ok := registry[scheme]
	return source, ok
}
