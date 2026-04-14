package cron

import (
	"errors"

	"github.com/robfig/cron/v3"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
)

func NewCron(cmd func(), opts ...Option) error {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		if errors.Is(err, conf.ErrInvalidKey) {
			log.Warn("job cron use default config!")
		} else {
			return err
		}
	}

	if cmd == nil {
		return errors.New("cron job function is nil")
	}
	spec := opt.config.Spec
	if spec == "" {
		return errors.New("cron spec is empty, please check config or use WithSpec")
	}

	if len(opt.cronOpts) == 0 {
		opt.cronOpts = append(opt.cronOpts, cron.WithSeconds())
	}

	c := &Cron{
		Cron: cron.New(opt.cronOpts...),
		opts: opt,
		cmd:  cmd,
	}
	if err := c.run(spec); err != nil {
		return err
	}

	hooks.Register(hooks.Stage_AfterStop, func() {
		c.Stop()
	})

	return nil
}
