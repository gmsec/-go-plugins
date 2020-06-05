module github.com/gmsec/goplugins

go 1.14

require (
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/gmsec/micro v0.0.0-20200605073327-6e725a16b2af
	github.com/google/uuid v1.1.1
	github.com/xxjwxc/public v0.0.0-20200604090416-5afd146414d5
	google.golang.org/grpc v1.29.1
)

// replace github.com/gmsec/micro => ../micro
