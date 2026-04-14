package etcdv3Registry

import "time"

type Config struct {
	ReadTimeout time.Duration
	ServiceTTL  int
	Prefix      string
}
