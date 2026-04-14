package metric

type Config struct {
	NameSpace string
	SubSystem string
}

func DefaultConfig() *Config {
	return &Config{
		NameSpace: "goxf",
		SubSystem: "ServerName",
	}
}

// Opts ...
type Opts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}
