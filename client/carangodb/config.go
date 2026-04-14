package carangodb

import (
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
)

type Config struct {
	// Endpoints holds 1 or more URL's used to connect to the database.
	// In case of a connection to an ArangoDB cluster, you must provide the URL's of all coordinators.
	Endpoints []string
	// DontFollowRedirect; if set, redirect will not be followed, api from the initial request will be returned without an error
	// DontFollowRedirect takes precendance over FailOnRedirect.
	DontFollowRedirect bool
	// FailOnRedirect; if set, redirect will not be followed, instead the status code is returned as error
	FailOnRedirect bool
	// DefaultTimeout is the timeout used by requests that have no timeout set in the given context.
	DefaultTimeout time.Duration
	// ConnLimit is the upper limit to the number of connections to a single server.
	// The default is 32 (DefaultConnLimit).
	// Set this value to -1 if you do not want any upper limit.
	ConnLimit int

	UserName string
	PassWord string
	DbName   string
}

type clientOption struct {
	config *Config

	transport http.RoundTripper
	// contentType specified type of content encoding to use.
	contentType driver.ContentType

	confPrefix string
	confName   string
}

type Option func(*clientOption)

func WithTransport(transport http.RoundTripper) Option {
	return func(o *clientOption) {
		o.transport = transport
	}
}
func WithContentType(contentType driver.ContentType) Option {
	return func(o *clientOption) {
		o.contentType = contentType
	}
}

func WithConfPrefix(prefix string) Option {
	return func(o *clientOption) {
		o.confPrefix = prefix
	}
}

func WithConfName(name string) Option {
	return func(o *clientOption) {
		o.confName = name
	}
}

func defaultConfig() *Config {
	return &Config{}
}

func defaultOptions() *clientOption {
	return &clientOption{
		config:     defaultConfig(),
		confPrefix: "client.arangodb",
		confName:   "default",
	}
}
