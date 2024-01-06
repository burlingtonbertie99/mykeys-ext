package server_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys/http"
	"github.com/stretchr/testify/require"
)

func TestAdminCheck(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	env := newEnv(t)
	srv := newTestServerEnv(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, alice, "alice", "github", env.client, clock)

	// PUT /sigchain/:id/:seq
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// POST /admin/check/:kid
	req, err = http.NewAuthRequest("POST", "/admin/check/"+alice.ID().String(), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"not authorized"}}`, string(body))

	// POST /admin/check/all
	req, err = http.NewAuthRequest("POST", "/admin/check/"+alice.ID().String(), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"not authorized"}}`, string(body))

	// Add admin
	srv.Server.SetAdmins([]keys.ID{bob.ID()})

	// POST /admin/check/:kid
	req, err = http.NewAuthRequest("POST", "/admin/check/"+alice.ID().String(), nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// POST /admin/check/all
	req, err = http.NewAuthRequest("POST", "/admin/check/all", nil, "", clock.Now(), bob)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))
}
