module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200220193200-0db28e999cf3
	github.com/keys-pub/keysd/http/api v0.0.0-20200220193726-e181801dd20a
	github.com/keys-pub/keysd/http/server v0.0.0-20200220193815-ae8034a06fbe
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
