package etcdv3

import (
	"fmt"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdv3Naming(t *testing.T) {
	re := NewEtcdv3NamingRegistry(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 3,
	})
	fmt.Println(re)
}
