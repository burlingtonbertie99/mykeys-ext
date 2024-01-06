module github.com/burlingtonbertie99/mykeys-ext/http/server

go 1.14

require (
	github.com/burlingtonbertie99/mykeys v0.0.1
	github.com/burlingtonbertie99/mykeys-ext/firestore v0.0.0-00010101000000-000000000000
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.1
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/keys-pub/keys v0.1.22 // indirect
	github.com/labstack/echo/v4 v4.2.1
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.43.0
)

replace github.com/burlingtonbertie99/mykeys => ../../../mykeys

replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../api

replace github.com/burlingtonbertie99/mykeys-ext/firestore => ../../firestore

replace github.com/keys-pub/vault => ../../../vault

replace github.com/burlingtonbertie99/mykeys-ext/ws/api => ../../ws/api
