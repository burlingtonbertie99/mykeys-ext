package server_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/http/server"
	"github.com/burlingtonbertie99/mykeys/dstore"
	"github.com/burlingtonbertie99/mykeys/encoding"
	"github.com/burlingtonbertie99/mykeys/http"
	"github.com/burlingtonbertie99/mykeys/tsutil"
	"github.com/burlingtonbertie99/mykeys/user"
	"github.com/stretchr/testify/require"
)

type testServerEnv struct {
	Server  *server.Server
	Handler http.Handler
	Emailer *testEmailer
	// Addr if started
	Addr string
}

// func testFirestore(t *testing.T) Fire {
// 	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
// 	fs, fsErr := firestore.New("firestore://chilltest-3297b", opts...)
// 	require.NoError(t, fsErr)
// 	err := fs.Delete(context.TODO(), "/")
// 	require.NoError(t, err)
// 	return fs
// }

func testFire(t *testing.T, clock tsutil.Clock) server.Fire {
	fi := dstore.NewMem()
	fi.SetClock(clock)
	fi.SetMode(dstore.FirestoreCompatibilityMode)
	return fi
}

func TestFireCreatedAt(t *testing.T) {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)

	err := fi.Set(context.TODO(), "/test/a", dstore.Data([]byte{0x01}))
	require.NoError(t, err)

	doc, err := fi.Get(context.TODO(), "/test/a")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ftime := doc.CreatedAt.Format(http.TimeFormat)
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", ftime)
	ftime = doc.CreatedAt.Format(tsutil.RFC3339Milli)
	require.Equal(t, "2009-02-13T23:31:30.001Z", ftime)
}

type env struct {
	clock    tsutil.Clock
	fi       server.Fire
	client   http.Client
	logLevel server.LogLevel
}

func newEnv(t *testing.T) *env {
	clock := tsutil.NewTestClock()
	fi := testFire(t, clock)
	return newEnvWithFire(t, fi, clock)
}

func newEnvWithFire(t *testing.T, fi server.Fire, clock tsutil.Clock) *env {
	client := http.NewClient()
	return &env{
		clock:    clock,
		fi:       fi,
		client:   client,
		logLevel: server.NoLevel,
	}
}

func newTestServerEnv(t *testing.T, env *env) *testServerEnv {
	rds := server.NewRedisTest(env.clock)
	srv := server.New(env.fi, rds, env.client, env.clock, server.NewLogger(env.logLevel))
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth(encoding.MustEncode(keys.RandBytes(32), encoding.Base62))
	err := srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")
	require.NoError(t, err)
	err = srv.SetTokenKey("f41deca7f9ef4f82e53cd7351a90bc370e2bf15ed74d147226439cfde740ac18")
	require.NoError(t, err)
	emailer := newTestEmailer()
	srv.SetEmailer(emailer)
	handler := server.NewHandler(srv)
	return &testServerEnv{
		Server:  srv,
		Handler: handler,
		Emailer: emailer,
	}
}

func (s *testServerEnv) Serve(req *http.Request) (int, nethttp.Header, []byte) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.Bytes()
}

type testEmailer struct {
	sentVerificationEmail map[string]string
}

func newTestEmailer() *testEmailer {
	return &testEmailer{sentVerificationEmail: map[string]string{}}
}

func (t *testEmailer) SentVerificationEmail(email string) string {
	s := t.sentVerificationEmail[email]
	return s
}

func (t *testEmailer) SendVerificationEmail(email string, code string) error {
	t.sentVerificationEmail[email] = code
	return nil
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}

func userMock(t *testing.T, key *keys.EdX25519Key, name string, service string, client http.Client, clock tsutil.Clock) *keys.Statement {
	url := ""
	api := ""

	id := hex.EncodeToString(sha256.New().Sum([]byte(service + "/" + name))[:8])

	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/"+id, name)
		api = "https://api.github.com/gists/" + id
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.ID())
	usr, err := user.New(key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := user.NewSigchainStatement(sc, usr, key, clock.Now())
	require.NoError(t, err)

	msg, err := usr.Sign(key)
	require.NoError(t, err)
	client.SetProxy(api, func(ctx context.Context, req *http.Request) http.ProxyResponse {
		return http.ProxyResponse{Body: []byte(githubMock(name, "1", msg))}
	})

	return st
}

func githubMock(name string, id string, msg string) string {
	msg = strings.ReplaceAll(msg, "\n", "")
	return `{
		"id": "` + id + `",
		"files": {
			"gistfile1.txt": {
				"content": "` + msg + `"
			}		  
		},
		"owner": {
			"login": "` + name + `"
		}
	  }`
}

func TestInternalAuth(t *testing.T) {
	env := newEnv(t)
	srv := newTestServerEnv(t, env)

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// POST /task/check/:kid
	req, err := http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"no auth token specified"}}`, string(body))

	// Set internal auth token
	srv.Server.SetInternalAuth("testtoken")

	// POST /task/check/:kid (with auth)
	req, err = http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "testtoken")
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", string(body))
}

func TestSpew(t *testing.T) {
	// To avoid import warning when we use spew
	spew.Sdump("testing")
}
