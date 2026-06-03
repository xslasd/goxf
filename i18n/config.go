package i18n

import "github.com/xslasd/goxf/conf"

type Config struct {
	Language string
	Path     string
	Exts     []string
}

type options struct {
	config     *Config
	confPrefix string
	confName   string

	unmarshal    conf.Unmarshal
	keyDelimiter string
}
type Option func(*options)

func WithConfPrefix(prefix string) Option {
	return func(o *options) {
		o.confPrefix = prefix
	}
}
func WithUnmarshal(unmarshal conf.Unmarshal) Option {
	return func(o *options) {
		o.unmarshal = unmarshal
	}
}

func WithKeyDelimiter(keyDelimiter string) Option {
	return func(o *options) {
		o.keyDelimiter = keyDelimiter
	}
}

func defaultConfig() *Config {
	return &Config{
		Language: "zh",
		Path:     "i18n",
		Exts:     []string{".yaml", ".yml", ".json", ".toml"},
	}
}

func defaultOptions() *options {
	return &options{
		config:       defaultConfig(),
		confPrefix:   "i18n",
		confName:     "default",
		keyDelimiter: ".",
	}
}
