package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gmsec/goplugins/registry/namingregister"
	"github.com/gmsec/micro/naming"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/xxjwxc/public/mylog"
	"github.com/xxjwxc/public/tools"
)

type namingClient struct {
	client naming_client.INamingClient
	//wch      etcd.WatchChan
	wch         chan string
	revision    int64
	closeed     bool
	serviceName string
}

// Put puts a key-value pair
func (nc *namingClient) Put(ctx context.Context, serviceName string, val naming.Update) error {
	ipPort, err := net.ResolveTCPAddr("tcp", val.Addr)
	if err != nil {
		return err
	}

	success, err := nc.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ipPort.IP.String(),
		Port:        uint64(ipPort.Port),
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"heart": fmt.Sprintf("%v", time.Now().Unix()),
			"op":    fmt.Sprintf("%v", val.Op),
		},
		// ClusterName: "cluster-a", // 默认值DEFAULT
		// GroupName:   "group-a",   // 默认值DEFAULT_GROUP
	})
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("nacos register error : addr:%v", val.Addr)
	}

	return nil
}

// Delete deletes a key, or optionally using WithRange(end), [key, end).
func (nc *namingClient) Delete(ctx context.Context, serviceName string, val naming.Update) error {
	ipPort, err := net.ResolveTCPAddr("tcp", val.Addr)
	if err != nil {
		return err
	}

	success, err := nc.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ipPort.IP.String(),
		Port:        uint64(ipPort.Port),
		ServiceName: serviceName,
		Ephemeral:   true,
		// Cluster:     "cluster-a", // 默认值DEFAULT
		// GroupName:   "group-a",   // 默认值DEFAULT_GROUP
	})
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("nacos register error : addr:%v", val.Addr)
	}

	return nil
}

// Get retrieves keys.
func (nc *namingClient) Get(ctx context.Context, serviceName string) ([]*naming.Update, error) {
	// SelectInstances 只返回满足这些条件的实例列表：healthy=${HealthyOnly},enable=true 和weight>0
	instances, err := nc.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		// GroupName:   "group-a",             // 默认值DEFAULT_GROUP
		// Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*naming.Update, 0, len(instances))
	for _, v := range instances {
		// string到int64
		i64, _ := strconv.ParseInt(v.Metadata["heart"], 10, 64)

		out = append(out, &naming.Update{
			Op:       naming.Add,
			Addr:     fmt.Sprintf("%v:%v", v.Ip, v.Port),
			Metadata: i64,
		})
	}

	return out, err
}

// Watchering Watcher is init
func (nc *namingClient) Watchering() bool {
	return nc.wch != nil
}

// Watch start watch // 做成全局监听
func (nc *namingClient) Watch(ctx context.Context, serviceName string) error {
	nc.wch = make(chan string)
	nc.closeed = false
	nc.serviceName = serviceName
	go func() {
		// 注意:我们可以在相同的key添加多个SubscribeCallback.
		err := nc.client.Subscribe(&vo.SubscribeParam{
			ServiceName: serviceName,
			// GroupName:   "group-a",             // 默认值DEFAULT_GROUP
			// Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
			SubscribeCallback: func(services []model.SubscribeService, err error) {
				log.Printf("\n\n callback return services:%v \n\n", tools.JSONDecode(services))
				var updates []*naming.Update
				for _, v := range services {
					i64, _ := strconv.ParseInt(v.Metadata["heart"], 10, 64)
					op := naming.Add
					if v.Metadata["op"] == "1" || !v.Enable || !v.Valid {
						op = naming.Delete
					}

					updates = append(updates, &naming.Update{
						Op:       op,
						Addr:     fmt.Sprintf("%v:%v", v.Ip, v.Port),
						Metadata: i64,
					})
				}
				if len(updates) > 0 && !nc.closeed {
					nc.wch <- tools.JSONDecode(updates)
				}
			},
		})

		if err != nil {
			mylog.Error(err)
		}
		// opts := []etcd.OpOption{etcd.WithRev(nc.revision + 1), etcd.WithPrefix(), etcd.WithPrevKV()}
		// nc.wch = nc.client.Watch(ctx, key, opts...)
	}()
	return nil
}

// WatcherNext watching next
func (nc *namingClient) WatcherNext() ([]*naming.Update, error) {
	jsonstr, closeed := <-nc.wch
	if closeed {
		nc.closeed = true
	}

	var updates []*naming.Update
	err := json.Unmarshal([]byte(jsonstr), &updates)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

// Close close
func (nc *namingClient) Close() error {
	nc.closeed = true
	if nc.wch != nil {
		close(nc.wch)
		return nc.client.Unsubscribe(&vo.SubscribeParam{
			ServiceName: nc.serviceName,
			// GroupName:   "group-a",             // 默认值DEFAULT_GROUP
			// Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
			SubscribeCallback: func(services []model.SubscribeService, err error) {
				log.Printf("\n\n callback return services:%s \n\n", tools.JSONDecode(services))
			},
		})
	}

	// return nc.client.Close()
	return nil
}

// New new watching client
func (nc *namingClient) New(serviceName string) namingregister.NamingClient {
	return &namingClient{
		client:      nc.client,
		serviceName: serviceName,
	}
}
