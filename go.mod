module github.com/gmsec/goplugins

go 1.14

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/gmsec/micro v0.0.0-20210823071050-448f4a058e29
	github.com/google/uuid v1.3.0
	github.com/gookit/color v1.4.2 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/soheilhy/cmux v0.1.4
	github.com/xxjwxc/public v0.0.0-20210812080902-893e9ff8ba5f
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210821163610-241b8fcbd6c8 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/examples v0.0.0-20210818220435-8ab16ef276a3 // indirect
)

// replace github.com/gmsec/micro => ../micro

// replace google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.29.1
