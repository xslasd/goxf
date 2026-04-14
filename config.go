package goxf

import (
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/flag"
)

type config struct {
	ServiceName    string
	EnableConsole  bool
	EnableTrace    bool
	EnableMetric   bool
	EnableRegister bool
	EnablePprof    bool
}
type Option func(*Service)

func WithAPPID(appID string) Option {
	return func(o *Service) {
		o.appID = appID
	}
}
func WithDefaultConfUnmarshal(unmarshal conf.Unmarshal) Option {
	return func(o *Service) {
		o.confUnmarshal = unmarshal
	}
}
func WithDefaultConfAddr(confAddr string) Option {
	return func(o *Service) {
		o.confAddr = confAddr
	}
}
func WithWatchConf(watch bool) Option {
	return func(o *Service) {
		o.isWatchConf = watch
	}
}
func WithLogger(key string) Option {
	return func(o *Service) {
		o.loggerKey = key
	}
}
func WithFlags(flags ...flag.Flag) Option {
	return func(o *Service) {
		flag.Register(flags...)
	}
}
func WithConfigPassword(password string) Option {
	return func(o *Service) {
		conf.SetPassword(password)
	}
}
