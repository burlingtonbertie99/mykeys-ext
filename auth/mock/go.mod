module github.com/burlingtonbertie99/mykeys-ext/auth/mock

go 1.14

require (
	github.com/alta/protopatch v0.3.3 // indirect
	github.com/burlingtonbertie99/mykeys v0.0.1
	github.com/burlingtonbertie99/mykeys-ext/auth/fido2 v0.0.1
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20210401141331-865547bb08e2 // indirect
)

replace github.com/burlingtonbertie99/mykeys-ext/auth/fido2 => ../fido2
