package plugin

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/xxjwxc/public/dev"
	"github.com/xxjwxc/public/mylog"

	"github.com/gin-gonic/gin"
	"github.com/gmsec/micro"
	"github.com/soheilhy/cmux"
)

// WithGin gin model init. gen 初始化模式
func WithGin(router *gin.Engine) Option {
	return func(o *options) {
		o.router = router
	}
}

// WithMicro micro model init. micro 初始化模式
func WithMicro(service micro.Service) Option {
	return func(o *options) {
		o.service = service
	}
}

// WithAddr addr model init. 地址初始化
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// Run run
func Run(opts ...Option) (*server, error) {
	var s server
	for _, f := range opts {
		f(&s.opt)
	}

	if s.opt.service == nil {
		return nil, fmt.Errorf("service is nil")
	}

	if len(s.opt.addr) > 0 {
		s.opt.service.Server().SetAddress(s.opt.addr)
	}

	// 起服务
	lis, err := net.Listen("tcp", s.opt.service.Server().GetAddress())
	if err != nil {
		mylog.Fatal("failed to listen: ", err)
		return nil, err
	}
	s.mux = cmux.New(lis)

	grpcl := s.mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	s.opt.service.Server().SetListener(grpcl)

	s.wg.Add(1)
	go func() { // grpc
		s.opt.service.Run()
		s.wg.Done()
	}()

	if s.opt.router != nil {
		s.wg.Add(1)
		go func() { // http
			// http.Handle("/", s.opt.router)
			// http.Serve(listener, nil)
			// or
			httpl := s.mux.Match(cmux.HTTP1Fast())
			err := s.opt.router.RunListener(httpl)
			if err != nil {
				debugPrintError(err)
			}
			s.wg.Done()
		}()
	}

	s.isStart = true
	return &s, nil
}

// Run run
// func Run(opts ...Option) (*server, error) {
// 	var s server
// 	for _, f := range opts {
// 		f(&s.opt)
// 	}

// 	if s.opt.service == nil {
// 		return nil, fmt.Errorf("service is nil")
// 	}

// 	if len(s.opt.addr) > 0 {
// 		s.opt.service.Server().SetAddress(s.opt.addr)
// 	}

// 	s.wg.Add(1)
// 	go func() { // grpc
// 		s.opt.service.Run()
// 		s.wg.Done()
// 	}()

// 	if s.opt.router != nil {
// 		s.wg.Add(1)
// 		listener := s.opt.service.Server().GetListener()
// 		go func() { // http
// 			http.Handle("/", s.opt.router)
// 			http.Serve(listener, nil)
// 			// or
// 			// err := s.opt.router.RunListener(listener)
// 			// if err != nil {
// 			// 	debugPrintError(err)
// 			// }
// 			s.wg.Done()
// 		}()
// 	}

// 	s.isStart = true
// 	return &s, nil
// }

// RunHTTP 只启动http
func RunHTTP(opts ...Option) (*server, error) {
	var s server
	for _, f := range opts {
		f(&s.opt)
	}

	if len(s.opt.addr) == 0 {
		return nil, fmt.Errorf("addr is nil")
	}

	if s.opt.router != nil {
		s.wg.Add(1)
		go func() { // http
			s.opt.router.Run(s.opt.addr)
			// http.Handle("/", s.opt.router)
			// http.Serve(listener, nil)
			// or
			// err := s.opt.router.RunListener(listener)
			// if err != nil {
			// 	debugPrintError(err)
			// }
			s.wg.Done()
		}()
	}

	s.isStart = true
	return &s, nil
}

func debugPrintError(err error) {
	if err != nil {
		if dev.IsDev() {
			fmt.Fprintf(os.Stderr, "[GIN-debug] [ERROR] %v\n", err)
		}
	}
}

type server struct {
	opt     options
	wg      sync.WaitGroup
	isStart bool
	mux     cmux.CMux
}

// Wait 等待结束
func (s *server) Wait() {
	if s.mux != nil {
		if err := s.mux.Serve(); err != nil {
			mylog.Error(err)
		}
	}

	s.wg.Wait()
}

// Stop 主动stop
func (s *server) Stop() {
	if !s.isStart {
		return
	}
	if s.opt.service != nil {
		go func() {
			s.opt.service.NotifyStop()
		}()
	}
	s.Wait()
}

type options struct {
	service micro.Service
	router  *gin.Engine
	addr    string
}

// Option option list
type Option func(*options)
