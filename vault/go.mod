module github.com/burlingtonbertie99/mykeys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/snappy v0.0.2 // indirect
	github.com/burlingtonbertie99/mykeys v0.1.22-0.20210523195800-d583c5244ce9
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.0-20210525002537-0c132efd0ef7
	github.com/burlingtonbertie99/mykeys-ext/http/client v0.0.0-20210525002537-0c132efd0ef7
	github.com/burlingtonbertie99/mykeys-ext/http/server v0.0.0-20210525002537-0c132efd0ef7
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
)

// replace github.com/burlingtonbertie99/mykeys => ../../keys

// replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../http/api

// replace github.com/burlingtonbertie99/mykeys-ext/http/client => ../http/client

// replace github.com/burlingtonbertie99/mykeys-ext/http/server => ../http/server
