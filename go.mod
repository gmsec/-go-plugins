module plugins

go 1.14

require (
	github.com/gin-gonic/gin v1.6.2
	github.com/gmsec/micro v0.0.0-20200509162038-ccfd45483427
	go.uber.org/zap v1.15.0
)

replace github.com/gmsec/micro => ../micro
