package carangodb

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/log"
)

func NewClient(opts ...Option) (*ArangodbCli, error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}
	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		return nil, err
	}

	cli, err := newClient(opt)
	if err != nil {
		return nil, err
	}
	log.Info("start arangoDB client ok.")
	return &ArangodbCli{
		cli,
		opt.config.DbName,
	}, nil
}

func NewClientDatabase(opts ...Option) (driver.Database, error) {
	cli, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}
	DB, err := cli.Database(context.Background(), cli.DBName)
	if err != nil {
		return nil, err
	}
	log.Info("start arangoDB client and connection database ok.")
	return DB, err
}
