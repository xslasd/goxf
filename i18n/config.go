package i18n

type Config struct {
	Language string
	Path     string
	Exts     []string
}

type options struct {
	config     *Config
	confPrefix string
	confName   string

	keyDelimiter string
}
type Option func(*options)

func WithConfPrefix(prefix string) Option {
	return func(o *options) {
		o.confPrefix = prefix
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
