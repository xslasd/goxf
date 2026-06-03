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

// NewSourceConf  创建配置文件实例
func NewSourceConf(configAddr string, confUnmarshal Unmarshal, watch bool) error {
	if configAddr == "" {
		return NoConfigErr
	}
	urlObj, err := url.Parse(configAddr)
	if err != nil {
		return err
	}
	var scheme = FileScheme
	if len(urlObj.Scheme) > 1 {
		scheme = urlObj.Scheme
	}
	creatorFunc, exist := GetDataSourceCreatorFunc(scheme)
	if !exist {
		return InvalidConfigSource
	}

	var ds ConfigSource

	switch scheme {
	case FileScheme:
		path := urlObj.Path
		// Windows 下 /D:/path -> D:/path
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		baseName := filepath.Base(absPath)
		ext := filepath.Ext(baseName)
		dir := filepath.Dir(absPath)
		isEnc := baseName == "system.enc"
		isLocal := strings.HasSuffix(baseName, ".local"+ext)

		if isEnc {
			watch = false
		}

		vPwd := verifyPassword(isEnc)

		wrapper := &fileSourceWrapper{
			isEnc:           isEnc,
			isLocal:         isLocal,
			vPwd:            vPwd,
			ext:             ext,
			absPath:         absPath,
			dir:             dir,
			customUnmarshal: confUnmarshal,
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

	if scheme == FileScheme {
		// FileScheme wrapper determines the final unmarshal format (e.g. after yaml merge)
		return LoadFromDataSource(ds, nil)
	}

	return LoadFromDataSource(ds, confUnmarshal)
}

func GetDataSourceCreatorFunc(scheme string) (DataSourceCreatorFunc, bool) {
	source, ok := registry[scheme]
	return source, ok
}
