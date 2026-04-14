package cron

import (
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
)

type Cron struct {
	*cron.Cron
	opts *options
	cmd  func()
}

func (c *Cron) run(spec string) error {
	name := c.opts.confPrefix + "." + c.opts.confName
	_, err := c.AddFunc(spec, func() {
		beg := time.Now()
		log.Debug("job start", log.String("name", name), log.String("spec", spec))
		code := "ok"
		defer func() {
			if err := recover(); err != nil {
				log.Error("job panic", log.String("name", name), log.Any("err", err))
				code = "panic"
			}
			log.Debug("job finish", log.String("name", name), log.FieldCost(time.Since(beg)))
			if c.opts.enableMetric {
				metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.JobType, name)
				metric.ServerHandleCounter.Inc(metric.JobType, name, code)
			}
		}()
		c.cmd()
	})
	if err != nil {
		return errors.Wrap(err, "cron add func error")
	}

	c.Start()
	log.Infof("job %s start ok, spec: %s", name, spec)
	return nil
}
