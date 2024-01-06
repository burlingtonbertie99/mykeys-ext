module github.com/burlingtonbertie99/mykeys-ext/wormhole

go 1.14

require (
	github.com/burlingtonbertie99/mykeys v0.1.20-0.20210102022201-ffb45798b8ab
	github.com/burlingtonbertie99/mykeys-ext/http/api v0.0.0-20210102023225-d2e7279d30fc
	github.com/burlingtonbertie99/mykeys-ext/http/client v0.0.0-20210102023902-79f9e8a3358a
	github.com/burlingtonbertie99/mykeys-ext/http/server v0.0.0-20210102023718-fd43795e6300
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gortc.io/stun v1.22.2
)

// replace github.com/burlingtonbertie99/mykeys => ../../keys

// replace github.com/burlingtonbertie99/mykeys-ext/http/api => ../http/api

// replace github.com/burlingtonbertie99/mykeys-ext/http/client => ../http/client

// replace github.com/burlingtonbertie99/mykeys-ext/http/server => ../http/server
