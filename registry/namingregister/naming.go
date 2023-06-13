package namingregister

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gmsec/micro/naming"
	"github.com/gmsec/micro/registry"
	"github.com/google/uuid"
	"github.com/xxjwxc/public/tools"
)

// NamingRegister dns default register
type NamingRegister struct {
	opts registry.Options
	sync.Mutex
	node    NamingClient
	address string
	// port    int

	// watch mabey
	// cancel context.CancelFunc
	// ctx    context.Context
	// isInit bool
	// listener
	// listener chan *mdns.ServiceEntry
}

func NewDNSNamingRegistry(nameCli NamingClient, opts ...registry.Option) registry.RegNaming {
	options := registry.Options{
		Context:          context.Background(),
		Timeout:          time.Millisecond * 100,
		NodeID:           uuid.New().String(),
		KeepHeartTimeout: time.Second * 30,
		ServiceName:      "gmsec.service",
	}
	for _, o := range opts {
		o(&options)
	}

	return &NamingRegister{
		opts: options,
		node: nameCli,
		// isInit: true,
	}
}

func (r *NamingRegister) String() string {
	return r.opts.ServiceName
}

func (r *NamingRegister) GetPort() int {
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
func (r *NamingRegister) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(&r.opts)
	}

	return nil
}

// Options get opts list
func (r *NamingRegister) Options() registry.Options {
	return r.opts
}

// Deregister 注销
func (r *NamingRegister) Deregister() error {
	r.Lock()
	defer r.Unlock()
	if r.node != nil {
		gr := &GRPCResolver{Client: r.node, HeartTimeout: r.opts.KeepHeartTimeout * 10}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		gr.Update(ctx, r.opts.ServiceName, naming.Update{
			Op:   naming.Delete,
			Addr: r.address,
			// Metadata: r.port,
		})
		cancel()
		r.node.Close()
		r.node = nil
	}

	return nil
}

// Register register & add new node
func (r *NamingRegister) Register(address string, Metadata interface{}) error {
	r.Lock()
	defer r.Unlock()
	// ip fix
	if strings.HasPrefix(address, "[::]") {
		address = tools.GetLocalIP() + address[4:]
	} else if strings.HasPrefix(address, ":") {
		address = tools.GetLocalIP() + address
	}
	r.address = address
	// ------end

	gr := &GRPCResolver{Client: r.node}
	//r.Resolve(fmt.Sprintf("127.0.0.1:%s", *port))
	up := naming.Update{
		Op:       naming.Add,
		Addr:     r.address,
		Metadata: time.Now().Unix(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := gr.Update(ctx, r.opts.ServiceName, up)
	cancel()
	// heart 心跳
	go func() {
		for {
			ticker := time.NewTicker(r.opts.KeepHeartTimeout)
			<-ticker.C
			up.Metadata = time.Now().Unix()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			gr.Update(ctx, r.opts.ServiceName, up) // 发送心跳
			cancel()
		}
	}()
	// ----------------------end

	return err
}

// Resolve resolve begin
func (r *NamingRegister) Resolve(target string) (naming.Watcher, error) {
	r.Lock()
	defer r.Unlock()

	t := &GRPCResolver{Client: r.node.New(target), HeartTimeout: r.opts.KeepHeartTimeout * 10}

	return t.Resolve(target)
}

// Close close watcher
func (r *NamingRegister) Close() { r.Deregister() }
