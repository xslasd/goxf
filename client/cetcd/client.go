package cetcd

import (
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"go.etcd.io/etcd/client/v3"
)

func newClient(opt *clientOption) (*clientv3.Client, error) {
	config := opt.config
	conf := clientv3.Config{
		Context:              opt.context,
		Endpoints:            config.Addrs,
		DialTimeout:          config.DialTimeout,
		AutoSyncInterval:     config.AutoSyncInterval,
		DialKeepAliveTime:    config.DialKeepAliveTime,
		DialKeepAliveTimeout: config.DialKeepAliveTimeout,
		MaxCallSendMsgSize:   config.MaxCallSendMsgSize,
		MaxCallRecvMsgSize:   config.MaxCallRecvMsgSize,
		TLS:                  opt.tls,
		Username:             config.Username,
		Password:             config.Password,
		RejectOldCluster:     config.RejectOldCluster,
		PermitWithoutStream:  config.PermitWithoutStream,
	}
	client, err := clientv3.New(conf)
	if err != nil {
		return nil, err
	}
	hooks.Register(hooks.Stage_AfterStop, func() {
		client.Close()
	})
	log.Info("start etcd client ok.")
	return client, nil
}
