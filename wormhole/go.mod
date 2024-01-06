module github.com/burlingtonbertie99/mykeys-ext/wormhole

go 1.14

require (
	github.com/burlingtonbertie99/mykeys v0.0.2
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.1
	github.com/burlingtonbertie99/mykeys-ext/http/client v0.0.1
	github.com/burlingtonbertie99/mykeys-ext/http/server v0.0.1
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	gortc.io/stun v1.22.2
)

//replace github.com/burlingtonbertie99/mykeys => ../../mykeys

replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../http/api

replace github.com/burlingtonbertie99/mykeys-ext/http/client => ../http/client

replace github.com/burlingtonbertie99/mykeys-ext/http/server => ../http/server
