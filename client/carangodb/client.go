package carangodb

import (
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/cluster"
	"github.com/arangodb/go-driver/http"
)

type ArangodbCli struct {
	driver.Client
	DBName string
}

func newClient(opt *clientOption) (driver.Client, error) {
	// Open a client connection
	config := opt.config
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints:          config.Endpoints,
		Transport:          opt.transport,
		DontFollowRedirect: config.DontFollowRedirect,
		FailOnRedirect:     config.FailOnRedirect,
		ConnectionConfig: cluster.ConnectionConfig{
			DefaultTimeout: config.DefaultTimeout,
		},
		ContentType: opt.contentType,
		ConnLimit:   config.ConnLimit,
	})
	if err != nil {
		return nil, err
	}
	// Client object
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.UserName, config.PassWord),
	})
	return client, err
}
