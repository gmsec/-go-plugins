# goplugins
Community maintained plugins for Go gmsec

gmsec plugins

### etcd
```
clientv3 "go.etcd.io/etcd/client/v3"
"github.com/gmsec/goplugins/registry/etcdv3"

	reg := etcdv3.NewEtcdv3NamingRegistry(clientv3.Config{
		Endpoints:   config.GetEtcdInfo().Addrs,
		DialTimeout: time.Second * time.Duration(config.GetEtcdInfo().Timeout),
	})
```
### nacos
```
"github.com/gmsec/goplugins/registry/nacos"

	nacosCnf := config.GetNacosNamingInfo()
	var serverconfig []constant.ServerConfig
	for _, v := range nacosCnf.Addrs {
		ipPort, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			panic(err)
		}
		serverconfig = append(serverconfig, *constant.NewServerConfig(
			ipPort.IP.String(),
			uint64(ipPort.Port),
		))
	}
	reg := nacos.NewNacosNamingRegistry(serverconfig, constant.NewClientConfig(constant.WithNamespaceId(nacosCnf.Namespace)))

```