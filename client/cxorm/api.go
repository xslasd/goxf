package cxorm

import (
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"xorm.io/xorm"
)

func NewClient(opts ...Option) (*xorm.EngineGroup, error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		return nil, err
	}

	return newClient(opt)
}
