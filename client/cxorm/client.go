package cxorm

import (
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"xorm.io/xorm"
)

func newClient(opt *clientOption) (*xorm.EngineGroup, error) {
	config := opt.config
	engine, err := xorm.NewEngineGroup(config.Driver, config.DataSources)
	if err != nil {
		return nil, err
	}
	logger, err := log.GetXormLogger(config.IsShowSQL)
	if err != nil {
		return nil, err
	}
	engine.SetLogger(logger)
	err = engine.Ping()
	if err != nil {
		return nil, err
	}
	hooks.Register(hooks.Stage_AfterStop, func() {
		engine.Close()
	})

	log.Info("start xorm client ok.")
	return engine, nil
}
