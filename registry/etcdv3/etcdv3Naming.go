package etcdv3

import (
	"fmt"
	"os"

	"github.com/gmsec/goplugins/registry/namingregister"
	"github.com/gmsec/micro/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewEtcdv3NamingRegistry returns a new default dns naming registry which is mdns
func NewEtcdv3NamingRegistry(conf clientv3.Config, opts ...registry.Option) registry.RegNaming {
	etcdCli, err := clientv3.New(conf)
	if err != nil {
		fmt.Printf("连接 etcd 服务器失败: %+v\n", err)
		os.Exit(1)
	}

	cli := &namingClient{client: etcdCli}

	return namingregister.NewDNSNamingRegistry(cli, opts...)
}
