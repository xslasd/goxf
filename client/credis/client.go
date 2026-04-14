package credis

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
)

func newClient(opt *clientOption) (*redis.Client, error) {
	config := opt.config
	if len(config.Addrs) == 0 {
		return nil, fmt.Errorf("addrs is empty")
	}
	ropt := redis.Options{
		Network:            config.Network,
		Addr:               config.Addrs[0],
		Username:           config.Username,
		Password:           config.Password,
		DB:                 config.DB,
		MaxRetries:         config.MaxRetries,
		MinRetryBackoff:    config.MinRetryBackoff,
		MaxRetryBackoff:    config.MaxRetryBackoff,
		DialTimeout:        config.DialTimeout,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		PoolSize:           config.PoolSize,
		MinIdleConns:       config.MinIdleConns,
		MaxConnAge:         config.MaxConnAge,
		PoolTimeout:        config.PoolTimeout,
		IdleTimeout:        config.IdleTimeout,
		IdleCheckFrequency: config.IdleCheckFrequency,
		TLSConfig:          opt.tls,
	}
	client := redis.NewClient(&ropt)
	_, err := client.Ping(opt.context).Result()
	if err != nil {
		return nil, err
	}
	hooks.Register(hooks.Stage_AfterStop, func() {
		client.Close()
	})
	log.Info("start redis client ok.")
	return client, nil

}

func newClusterClient(opt *clientOption) (*redis.ClusterClient, error) {
	config := opt.config
	copt := redis.ClusterOptions{
		Addrs:              config.Addrs,
		Username:           config.Username,
		Password:           config.Password,
		MaxRetries:         config.MaxRetries,
		MinRetryBackoff:    config.MinRetryBackoff,
		MaxRetryBackoff:    config.MaxRetryBackoff,
		DialTimeout:        config.DialTimeout,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		PoolSize:           config.PoolSize,
		MinIdleConns:       config.MinIdleConns,
		MaxConnAge:         config.MaxConnAge,
		PoolTimeout:        config.PoolTimeout,
		IdleTimeout:        config.IdleTimeout,
		IdleCheckFrequency: config.IdleCheckFrequency,
		TLSConfig:          opt.tls,
	}
	clusterClient := redis.NewClusterClient(&copt)
	_, err := clusterClient.Ping(opt.context).Result()
	if err != nil {
		return nil, err
	}
	hooks.Register(hooks.Stage_AfterStop, func() {
		clusterClient.Close()
	})
	log.Info("start redis client ok.")
	return clusterClient, err
}
