package etcdv3

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gmsec/micro/registry"
	"github.com/google/uuid"
	"github.com/xxjwxc/public/tools"
	"google.golang.org/grpc/naming"
)

// Etcdv3NamingRegister dns default register
type Etcdv3NamingRegister struct {
	opts registry.Options
	sync.Mutex
	node    *clientv3.Client
	address string
	// port    int

	// watch mabey
	// cancel context.CancelFunc
	// ctx    context.Context
	// isInit bool
	// listener
	// listener chan *mdns.ServiceEntry
}

// NewEtcdv3NamingRegistry returns a new default dns naming registry which is mdns
func NewEtcdv3NamingRegistry(conf clientv3.Config, opts ...registry.Option) registry.RegNaming {
	etcdCli, err := clientv3.New(conf)
	if err != nil {
		fmt.Printf("连接 etcd 服务器失败: %+v\n", err)
		os.Exit(1)
	}

	return newDNSNamingRegistry(etcdCli, opts...)
}

func newDNSNamingRegistry(etcdCli *clientv3.Client, opts ...registry.Option) registry.RegNaming {
	options := registry.Options{
		Context:          context.Background(),
		Timeout:          time.Millisecond * 100,
		NodeID:           uuid.New().String(),
		KeepHeartTimeout: time.Second * 15,
		ServiceName:      "gmsec.service",
	}
	for _, o := range opts {
		o(&options)
	}

	return &Etcdv3NamingRegister{
		opts: options,
		node: etcdCli,
		// isInit: true,
	}
}

func (r *Etcdv3NamingRegister) String() string {
	return r.opts.ServiceName
}

func (r *Etcdv3NamingRegister) GetPort() int {
	if len(r.opts.Addrs) == 0 {
		return 0
	}

	_, pt, err := net.SplitHostPort(r.opts.Addrs[0])
	if err != nil {
		return 0
	}
	port, _ := strconv.Atoi(pt)
	return port
}

// Init init option
func (r *Etcdv3NamingRegister) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(&r.opts)
	}

	return nil
}

// Options get opts list
func (r *Etcdv3NamingRegister) Options() registry.Options {
	return r.opts
}

// Deregister 注销
func (r *Etcdv3NamingRegister) Deregister() error {
	r.Lock()
	defer r.Unlock()
	if r.node != nil {
		gr := &GRPCResolver{Client: r.node, HeartTimeout: r.opts.KeepHeartTimeout * 10}
		gr.Update(context.TODO(), r.opts.ServiceName, naming.Update{
			Op:   naming.Delete,
			Addr: r.address,
			// Metadata: r.port,
		})
		r.node.Close()
		r.node = nil
	}

	return nil
}

//Register register & add new node
func (r *Etcdv3NamingRegister) Register(address string, Metadata interface{}) error {
	r.Lock()
	defer r.Unlock()
	// ip fix
	if strings.HasPrefix(address, "[::]") {
		address = tools.GetLocalIP() + address[4:]
	} else if strings.HasPrefix(address, ":") {
		address = tools.GetLocalIP() + address
	}
	// ------end

	gr := &GRPCResolver{Client: r.node}
	//r.Resolve(fmt.Sprintf("127.0.0.1:%s", *port))
	up := naming.Update{
		Op:       naming.Add,
		Addr:     address,
		Metadata: time.Now().Unix(),
	}
	err := gr.Update(context.TODO(), r.opts.ServiceName, up)

	// heart 心跳
	go func() {
		for {
			ticker := time.NewTicker(r.opts.KeepHeartTimeout)
			<-ticker.C
			up.Metadata = time.Now().Unix()
			gr.Update(context.TODO(), r.opts.ServiceName, up) // 发送心跳
		}
	}()
	// ----------------------end

	return err
}

// Resolve resolve begin
func (r *Etcdv3NamingRegister) Resolve(target string) (naming.Watcher, error) {
	r.Lock()
	defer r.Unlock()

	t := &GRPCResolver{Client: r.node, HeartTimeout: r.opts.KeepHeartTimeout * 10}

	return t.Resolve(target)
}

// Close close watcher
func (r *Etcdv3NamingRegister) Close() { r.Deregister() }
