module github.com/gmsec/goplugins

go 1.14

require (
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/gmsec/micro v0.0.0-20200512170207-8dff54fc0fa7
	github.com/google/uuid v1.1.1
	github.com/micro/mdns v0.3.0
	github.com/xxjwxc/public v0.0.0-20200603120633-8d5885879a2f
	google.golang.org/grpc v1.29.1
)

replace github.com/gmsec/micro => ../micro
