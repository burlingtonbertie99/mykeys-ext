module github.com/burlingtonbertie99/mykeys-ext/http/client

go 1.14

require (
	github.com/burlingtonbertie99/mykeys v0.0.4
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.1
	github.com/burlingtonbertie99/mykeys-ext/http/server v0.0.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
)

replace github.com/burlingtonbertie99/mykeys => ../../../mykeys

replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../api

replace github.com/burlingtonbertie99/mykeys-ext/http/server => ../server

replace github.com/keys-pub/vault => ../../../vault

replace github.com/burlingtonbertie99/mykeys-ext/firestore => ../../firestore

replace github.com/burlingtonbertie99/mykeys-ext/ws/api => ../../ws/api
