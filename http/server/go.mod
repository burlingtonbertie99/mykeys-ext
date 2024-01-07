module github.com/burlingtonbertie99/mykeys-ext/http/server

go 1.14

require (
	github.com/burlingtonbertie99/mykeys v0.0.4
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.1
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo/v4 v4.2.1
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/net v0.0.0-20210326060303-6b1517762897 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6 // indirect
	golang.org/x/text v0.3.5 // indirect
)

replace github.com/burlingtonbertie99/mykeys => ../../../mykeys

replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../api

replace github.com/burlingtonbertie99/mykeys-ext/firestore => ../../firestore

//replace github.com/keys-pub/vault => ../../../vault

replace github.com/burlingtonbertie99/mykeys-ext/ws/api => ../../ws/api
