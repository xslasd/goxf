package conf

import (
	"encoding/json"
	"io"
	"net/url"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type ConfigSource interface {
	ReadConfig() ([]byte, error) //读取配置内容
	Changed() <-chan struct{}    //监听配置是否改动
	io.Closer                    //io 关闭
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

// NewDataSource ..
func NewDataSource(configAddr string, watch bool) (ConfigSource, error) {
	if configAddr == "" {
		return nil, NoConfigErr
	}
	urlObj, err := url.Parse(configAddr)
	if err != nil {
		return nil, err
	}
	var scheme = "file"
	if len(urlObj.Scheme) > 1 {
		scheme = urlObj.Scheme
	}

	creatorFunc, exist := GetDataSourceCreatorFunc(scheme)
	if !exist {
		return nil, InvalidConfigSource
	}
	return creatorFunc(configAddr, watch), nil
}

func GetDataSourceCreatorFunc(scheme string) (DataSourceCreatorFunc, bool) {
	source, ok := registry[scheme]
	return source, ok
}

func ExtToUnmarshal(path string) (Unmarshal, error) {
	var confUnmarshal Unmarshal
	switch filepath.Ext(path) {
	case ".toml":
		confUnmarshal = toml.Unmarshal
	case ".yaml", ".yml":
		confUnmarshal = yaml.Unmarshal
	case ".json":
		confUnmarshal = json.Unmarshal
	default:
		return nil, UnmarshalInvalid
	}
	return confUnmarshal, nil
}
