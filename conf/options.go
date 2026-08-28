package conf

// Options represents configuration options for creating or loading a Conf instance.
type Options struct {
	unmarshal         Unmarshal
	watch             bool
	keyDelimiter      string
	password          string
	passwordCryptFunc func(string) string
	useGlobalPassword bool
}

// Option defines a functional option to configure Options.
type Option func(*Options)

// defaultOptions returns the default configuration options.
func defaultOptions() *Options {
	return &Options{
		keyDelimiter:      ".",
		useGlobalPassword: true,
	}
}

func (o *Options) getEffectivePassword() (string, func(string) string) {
	if o.password != "" || o.passwordCryptFunc != nil {
		return o.password, o.passwordCryptFunc
	}
	if o.useGlobalPassword {
		return configPassword, passwordCryptFunc
	}
	return "", nil
}

// WithUnmarshal sets a custom Unmarshal function for parsing configuration content.
func WithUnmarshal(unmarshal Unmarshal) Option {
	return func(o *Options) {
		o.unmarshal = unmarshal
	}
}

// WithWatch enables or disables watching config changes for hot reload.
func WithWatch(watch bool) Option {
	return func(o *Options) {
		o.watch = watch
	}
}

// WithKeyDelimiter sets the delimiter used for nested keys (default is ".").
func WithKeyDelimiter(delimiter string) Option {
	return func(o *Options) {
		o.keyDelimiter = delimiter
	}
}

// WithPassword sets the password and optional crypt function for configuration decryption/encryption.
func WithPassword(pwd string, cryptFn ...func(string) string) Option {
	return func(o *Options) {
		o.password = pwd
		if len(cryptFn) > 0 {
			o.passwordCryptFunc = cryptFn[0]
		} else {
			o.passwordCryptFunc = nil
		}
	}
}

// WithIgnoreGlobalPassword disables inheriting global password/cryptFunc when not set in options.
func WithIgnoreGlobalPassword() Option {
	return func(o *Options) {
		o.useGlobalPassword = false
	}
}
