package server_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys/dstore"
	"github.com/stretchr/testify/require"
)

func TestSigchains(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	// firestore.SetContextLogger(NewContextLogger(DebugLevel))

	env := newEnv(t)
	srv := newTestServerEnv(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// GET /invalidloc (not found)
	req, err := http.NewRequest("GET", "/invalidloc", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	expected := `{"error":{"code":404,"message":"resource not found"}}`
	require.Equal(t, expected, string(body))

	// PUT /sigchains (method not allowed)
	req, err = http.NewRequest("PUT", "/sigchains", bytes.NewReader([]byte("test")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	expected = `{"error":{"code":405,"message":"method not allowed"}}`
	require.Equal(t, expected, string(body))

	// Alice sign "testing"
	sca := keys.NewSigchain(alice.ID())
	sta, err := keys.NewSigchainStatement(sca, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = sca.Add(sta)
	require.NoError(t, err)
	staBytes, err := sta.Bytes()
	require.NoError(t, err)

	// PUT /sigchain/:kid/:seq
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(staBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// PUT /sigchain/:kid/:seq again (conflict error)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(staBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	expected = `{"error":{"code":409,"message":"statement already exists"}}`
	require.Equal(t, expected, string(body))

	// Bob sign "testing"
	scb := keys.NewSigchain(bob.ID())
	stb, err := keys.NewSigchainStatement(scb, []byte("testing"), bob, "", clock.Now())
	require.NoError(t, err)

	// PUT /sigchain/:kid/:seq (invalid, bob's statement)
	b, err := stb.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader([]byte(b)))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"invalid kid"}}`
	require.Equal(t, expected, string(body))

	// PUT /sigchain/:kid/:seq (empty json)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader([]byte("{}")))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"not enough bytes for statement"}}`
	require.Equal(t, expected, string(body))

	// PUT /sigchain/:kid/:seq (no body)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected = `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, string(body))

	// GET /sigchain/:kid/:seq
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	code, header, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", header.Get("CreatedAt"))
	require.Equal(t, "2009-02-13T23:31:30.002Z", header.Get("CreatedAt-RFC3339M"))
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", header.Get("Last-Modified"))
	require.Equal(t, "2009-02-13T23:31:30.002Z", header.Get("Last-Modified-RFC3339M"))
	expectedSigned := `{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}`
	require.Equal(t, expectedSigned, string(body))

	// GET /sigchain/:kid
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain := `{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","statements":[{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain, string(body))

	// GET /sigchain/:kid (not found)
	req, err = http.NewRequest("GET", dstore.Path("sigchain", keys.RandID("kex")), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"sigchain not found"}}`, string(body))

	// GET /sigchain/:kid?include=md
	req, err = http.NewRequest("GET", fmt.Sprintf("/sigchain/%s?include=md", alice.ID()), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain2 := `{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","md":{"/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077/1":{"createdAt":"2009-02-13T23:31:30.002Z","updatedAt":"2009-02-13T23:31:30.002Z"}},"statements":[{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain2, string(body))

	// GET /sigchains
	// req, err = http.NewRequest("GET", "/sigchains", nil)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// require.Equal(t, http.StatusOK, code)
	// expectedSigs := `{"statements":[{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}],"version":"1234567890003"}`
	// require.Equal(t, expectedSigs, string(body))

	// GET /sigchains?include=md&limit=1
	// req, err = http.NewRequest("GET", "/sigchains?include=md&limit=1", nil)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// require.Equal(t, http.StatusOK, code)
	// expectedSigsWithMetadata := `{"md":{"/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077/1":{"createdAt":"2009-02-13T23:31:30.002Z","updatedAt":"2009-02-13T23:31:30.002Z"}},"statements":[{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}],"version":"1234567890003"}`
	// require.Equal(t, expectedSigsWithMetadata, string(body))

	// GET /:kid
	req, err = http.NewRequest("GET", dstore.Path(alice.ID()), nil)
	require.NoError(t, err)
	req.Host = "sigcha.in"
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedSigchain = `{"kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","statements":[{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}]}`
	require.Equal(t, expectedSigchain, string(body))

	// Alice sign "testing2"
	sta2, err := keys.NewSigchainStatement(sca, []byte("testing2"), alice, "", clock.Now())
	require.NoError(t, err)
	err = sca.Add(sta2)
	require.NoError(t, err)

	// GET /:kid/:seq
	req, err = http.NewRequest("GET", fmt.Sprintf("/%s/1", alice.ID()), nil)
	require.NoError(t, err)
	req.Host = "sigcha.in"
	code, header, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", header.Get("CreatedAt"))
	require.Equal(t, "2009-02-13T23:31:30.002Z", header.Get("CreatedAt-RFC3339M"))
	require.Equal(t, "Fri, 13 Feb 2009 23:31:30 GMT", header.Get("Last-Modified"))
	require.Equal(t, "2009-02-13T23:31:30.002Z", header.Get("Last-Modified-RFC3339M"))
	expectedSigned = `{".sig":"j5FZVQKWrnclXHHHIVX7JZ0letgR22cGl7ItlAUHqEsW+kCCMZvDBGEunVJScjVphrqGrPb7oCuMZouGv7GwCQ==","data":"dGVzdGluZw==","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","seq":1,"ts":1234567890001}`
	require.Equal(t, expectedSigned, string(body))

	// PUT /sigchain/:kid/:seq
	b, err = sta2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/%s/2", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	req.Host = "sigcha.in"
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// PUT /sigchain/invalidloc/1
	req, err = http.NewRequest("PUT", dstore.Path("/sigchain/invalidloc", 1), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"invalid kid"}}`, string(body))

	// Alice sign large message
	large := bytes.Repeat([]byte{0x01}, 17*1024)
	sta, err = keys.NewSigchainStatement(sca, large, alice, "", clock.Now())
	require.NoError(t, err)
	err = sca.Add(sta)
	require.NoError(t, err)

	// PUT /sigchain/:kid/:seq (too large)
	b, err = sta.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/%d", alice.ID(), sta.Seq), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"too much data for sigchain statement (greater than 16KiB)"}}`, string(body))

	// GET /foo/bar
	req, err = http.NewRequest("GET", "/foo/bar", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, string(body))

	// GET /:kid/bar
	req, err = http.NewRequest("GET", dstore.Path(alice.ID(), "bar"), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"resource not found"}}`, string(body))
}
