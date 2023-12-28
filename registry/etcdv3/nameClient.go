package etcdv3

import (
	"context"
	"encoding/json"

	"github.com/gmsec/goplugins/registry/namingregister"
	"github.com/gmsec/micro/naming"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type namingClient struct {
	client   *clientv3.Client
	wch      etcd.WatchChan
	revision int64
}

// Put puts a key-value pair
func (nc *namingClient) Put(ctx context.Context, key string, val naming.Update) error {
	var v []byte
	var err error
	if v, err = json.Marshal(val); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = nc.client.Put(ctx, key+"//"+val.Addr, string(v))
	return err
}

// Delete deletes a key, or optionally using WithRange(end), [key, end).
func (nc *namingClient) Delete(ctx context.Context, key string, val naming.Update) error {
	_, err := nc.client.Delete(ctx, key+"//"+val.Addr)
	return err
}

// Get retrieves keys.
func (nc *namingClient) Get(ctx context.Context, key string) ([]*naming.Update, error) {
	resp, err := nc.client.Get(ctx, key+"//", etcd.WithPrefix(), etcd.WithSerializable())
	if err != nil {
		return nil, err
	}
	if resp.Header != nil {
		nc.revision = resp.Header.Revision
	}

	out := make([]*naming.Update, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var jupdate naming.Update
		if err := json.Unmarshal(kv.Value, &jupdate); err != nil {
			continue
		}
		out = append(out, &jupdate)
	}

	return out, err
}

// Watchering Watcher is init
func (nc *namingClient) Watchering() bool {
	return nc.wch != nil
}

// Watch start watch
func (nc *namingClient) Watch(ctx context.Context, key string) error {
	opts := []etcd.OpOption{etcd.WithPrefix(), etcd.WithPrevKV(), etcd.WithSerializable()}
	nc.wch = nc.client.Watch(ctx, key+"//", opts...)
	return nil
}

// WatcherNext watching next
func (nc *namingClient) WatcherNext() ([]*naming.Update, error) {
	wr, ok := <-nc.wch
	if !ok {
		return nil, status.Error(codes.Unavailable, namingregister.ErrWatcherClosed.Error())
	}
	if err := wr.Err(); err != nil {
		return nil, err
	}
	updates := make([]*naming.Update, 0, len(wr.Events))
	for _, e := range wr.Events {
		var jupdate naming.Update
		var err error
		switch e.Type {
		case etcd.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, &jupdate)
			jupdate.Op = naming.Add
		case etcd.EventTypeDelete:
			err = json.Unmarshal(e.PrevKv.Value, &jupdate)
			jupdate.Op = naming.Delete
		}

		if err == nil {
			updates = append(updates, &jupdate)
		}
	}
	return updates, nil
}

// New new watching client
func (nc *namingClient) New(serviceName string) namingregister.NamingClient {
	return &namingClient{
		client: nc.client,
	}
}

// Close close
func (nc *namingClient) Close() error {
	// return nc.client.Close()
	return nil
}
