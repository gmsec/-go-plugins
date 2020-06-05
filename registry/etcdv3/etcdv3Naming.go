package etcdv3

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdNaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/gmsec/micro/registry"
	"github.com/google/uuid"
	"google.golang.org/grpc/naming"
)

// Etcdv3NamingRegister dns default register
type Etcdv3NamingRegister struct {
	opts registry.Options
	sync.Mutex
	node *clientv3.Client

	// watch mabey
	cancel context.CancelFunc
	ctx    context.Context
	isInit bool
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
		Context:     context.Background(),
		Timeout:     time.Millisecond * 100,
		NodeID:      uuid.New().String(),
		ServiceName: "gmsec.service",
	}
	for _, o := range opts {
		o(&options)
	}

	return &Etcdv3NamingRegister{
		opts:   options,
		node:   etcdCli,
		isInit: true,
	}
}

func (r *Etcdv3NamingRegister) String() string {
	return r.opts.ServiceName
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
		r.node.Close()
		r.node = nil
	}

	return nil
}

//Register register & add new node
func (r *Etcdv3NamingRegister) Register(address string, Metadata interface{}) error {
	r.Lock()
	defer r.Unlock()

	r.opts.Addrs = []string{address}
	_, pt, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	port, _ := strconv.Atoi(pt)

	gr := &etcdNaming.GRPCResolver{Client: r.node}

	//r.Resolve(fmt.Sprintf("127.0.0.1:%s", *port))
	err = gr.Update(context.TODO(), r.opts.ServiceName, naming.Update{
		Op:       naming.Add,
		Addr:     address,
		Metadata: port,
	})

	return err
}

// Resolve resolve begin
func (r *Etcdv3NamingRegister) Resolve(target string) (naming.Watcher, error) {
	r.Lock()
	defer r.Unlock()

	t := &etcdNaming.GRPCResolver{Client: r.node}

	return t.Resolve(target)
}

// Close close watcher
func (r *Etcdv3NamingRegister) Close() { r.node.Close() }
