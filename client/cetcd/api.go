package cetcd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewClient(opts ...Option) (*clientv3.Client, error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		return nil, err
	}

	if opt.config.CaCert != "" && opt.config.CertFile != "" && opt.config.KeyFile != "" {
		certBytes, err := os.ReadFile(opt.config.CaCert)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)
		if !ok {
			return nil, fmt.Errorf("解析 PEM 编码证书失败")
		}
		tlsConfig.RootCAs = caCertPool
		tlsCert, err := tls.LoadX509KeyPair(opt.config.CertFile, opt.config.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		opt.tls = tlsConfig
	}
	return newClient(opt)
}
