package etcdv3Registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/registry"
	"github.com/xslasd/goxf/server"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type EtcdRegistry struct {
	name string
	sync.RWMutex
	sessions map[string]*concurrency.Session
	client   *clientv3.Client
	config   *Config
}

func NewRegistry(client *clientv3.Client, config *Config) *EtcdRegistry {
	return &EtcdRegistry{
		name:     "etcd",
		client:   client,
		config:   config,
		sessions: make(map[string]*concurrency.Session, 0),
	}
}

const registerRetryRate = 1

func (e *EtcdRegistry) Register(ctx context.Context, info *server.ServiceInfo) error {
	key := registry.GetServiceKey(e.config.Prefix, info)
	val := registry.GetServiceValue(info)
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, e.config.ReadTimeout)
		defer cancel()
	}
	opOptions := make([]clientv3.OpOption, 0)
	session, err := e.getSession(key, concurrency.WithTTL(e.config.ServiceTTL))
	if err != nil {
		log.Warn("register service getSession", log.FieldErr(err), log.FieldKeyAny(key))
		return err
	}
	opOptions = append(opOptions, clientv3.WithLease(session.Lease()))

	_, err = e.client.Put(ctx, key, val, opOptions...)
	if err != nil {
		log.Warn("register service", log.FieldErr(err), log.FieldKeyAny(key))
		return err
	}
	log.Infof("register %s service: %s access address[etcd:/%s]", info.Scheme, info.Name, key)
	return nil
}

func (e *EtcdRegistry) UnRegister(ctx context.Context, info *server.ServiceInfo) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.config.ReadTimeout)
		defer cancel()
	}
	key := registry.GetServiceKey(e.config.Prefix, info)
	if err := e.delSession(key); err != nil {
		return err
	}
	_, err := e.client.Delete(ctx, key)
	return err
}

func (e *EtcdRegistry) ListServices(ctx context.Context, s string) ([]*server.ServiceInfo, error) {
	fmt.Println("====ListServices=====")
	panic("implement me")
}

func (e *EtcdRegistry) Watch(ctx context.Context, name string) (chan *registry.Endpoints, error) {
	prefix := fmt.Sprintf("/%s%s", e.config.Prefix, name)
	addresses := make(chan *registry.Endpoints, 10)
	ept := &registry.Endpoints{
		Nodes: make(map[string]server.ServiceInfo),
	}
	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	updateAddrList(ept, prefix, resp.Kvs...)
	addresses <- ept
	go func() {
		rch := e.client.Watch(ctx, prefix, clientv3.WithPrefix())
		for n := range rch {
			for _, ev := range n.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					updateAddrList(ept, prefix, ev.Kv)
				case clientv3.EventTypeDelete:
					deleteAddrList(ept, prefix, ev.Kv)
				}
				addresses <- ept
			}
		}
	}()
	return addresses, nil
}

func (e *EtcdRegistry) Kind() string {
	return e.name
}

func (e *EtcdRegistry) getSession(k string, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	e.RLock()
	sess, ok := e.sessions[k]
	e.RUnlock()
	if ok {
		return sess, nil
	}
	sess, err := concurrency.NewSession(e.client, opts...)
	if err != nil {
		return sess, err
	}
	e.Lock()
	e.sessions[k] = sess
	e.Unlock()
	return sess, nil
}

func (e *EtcdRegistry) delSession(k string) error {
	if e.config.ServiceTTL > 0 {
		e.RLock()
		sess, ok := e.sessions[k]
		e.RUnlock()
		if ok {
			e.Lock()
			delete(e.sessions, k)
			e.Unlock()
			if err := sess.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateAddrList(al *registry.Endpoints, prefix string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix+"://")
		if addr == "" {
			continue
		}
		var serviceInfo server.ServiceInfo
		if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
			log.Error("parse uri", log.FieldErr(err), log.FieldKey(string(kv.Key)))
			continue
		}
		al.Nodes[addr] = serviceInfo
	}
}

func deleteAddrList(al *registry.Endpoints, prefix string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix+"://")
		if addr == "" {
			continue
		}
		delete(al.Nodes, addr)
	}
}
