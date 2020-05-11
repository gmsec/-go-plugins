package ginplugins

import (
	"sync"

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
func Run(opts ...Option) (bool, *server) {
	var s server
	for _, f := range opts {
		f(&s.opt)
	}

	if s.opt.service == nil {
		return false, nil
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
			s.opt.router.RunListener(listener)
			s.wg.Done()
		}()
	}

	s.isStart = true
	return true, &s
}

type server struct {
	opt     options
	wg      sync.WaitGroup
	isStart bool
}

func (s *server) Stop() {
	if !s.isStart {
		return
	}
	if s.opt.service != nil {
		s.opt.service.NotifyStop()
	}
	if s.opt.router != nil {
		listener := s.opt.service.Server().GetListener()
		listener.Close()
	}
	s.wg.Wait()
}

type options struct {
	service micro.Service
	router  *gin.Engine
	addr    string
}

// Option option list
type Option func(*options)
