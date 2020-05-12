package plugin

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsec/micro"
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

	go func() {
		s.wg.Add(1)
		s.opt.service.Run()
		s.wg.Done()
	}()

	if s.opt.router != nil {
		listener := s.opt.service.Server().GetListener()
		go func() {
			s.wg.Add(1)
			err := s.opt.router.RunListener(listener)
			if err != nil {
				fmt.Println(err)
			}
			s.wg.Done()
		}()
	}

	s.isStart = true
	return &s, nil
}

type server struct {
	opt     options
	wg      sync.WaitGroup
	isStart bool
}

// Wait 等待结束
func (s *server) Wait() {
	time.Sleep(1 * time.Second)
	s.wg.Wait()
}

// Stop 主动stop
func (s *server) Stop() {
	if !s.isStart {
		return
	}
	if s.opt.service != nil {
		s.opt.service.NotifyStop()
	}
	// if s.opt.router != nil {
	// 	listener := s.opt.service.Server().GetListener()
	// 	listener.Close()
	// }
	s.Wait()
}

type options struct {
	service micro.Service
	router  *gin.Engine
	addr    string
}

// Option option list
type Option func(*options)
