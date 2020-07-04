module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200704210752-498c4412af12
	github.com/keys-pub/keys-ext/firestore v0.0.0-20200704211016-ce8ce10a1087
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200704211016-ce8ce10a1087
	github.com/labstack/echo/v4 v4.1.16
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/api v0.25.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
