package nacos

import (
	"fmt"
	"os"

	"github.com/gmsec/goplugins/registry/namingregister"
	"github.com/gmsec/micro/registry"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// NewNacosNamingRegistry returns a new default dns naming registry which is mdns
func NewNacosNamingRegistry(sconf []constant.ServerConfig, conf *constant.ClientConfig, opts ...registry.Option) registry.RegNaming {
	// a more graceful way to create naming client
	nacosClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  conf,
			ServerConfigs: sconf,
		},
	)
	if err != nil {
		fmt.Printf("连接 etcd 服务器失败: %+v\n", err)
		os.Exit(1)
	}

	cli := &namingClient{client: nacosClient}

	return namingregister.NewDNSNamingRegistry(cli, opts...)
}
