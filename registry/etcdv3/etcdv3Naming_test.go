package etcdv3

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdv3Naming(t *testing.T) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.155.1.121:30400"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		fmt.Printf("连接 etcd 服务器失败: %+v\n", err)
		os.Exit(1)
	}
	cli := &namingClient{client: etcdCli}
	list, _ := cli.Get(context.Background(), "haihuman.srv")
	fmt.Println(list)

	re := NewEtcdv3NamingRegistry(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 3,
	})
	fmt.Println(re)
}
