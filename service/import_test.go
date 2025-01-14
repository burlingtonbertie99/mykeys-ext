package service

import (
	"context"
	"testing"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/vault/keyring"
	"github.com/burlingtonbertie99/mykeys/api"
	"github.com/stretchr/testify/require"
)

func TestKeyImport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	key := keys.GenerateEdX25519Key()
	export, err := api.EncodeKey(api.NewKey(key), "testpassword")
	require.NoError(t, err)

	// Import
	importResp, err := service.KeyImport(ctx, &KeyImportRequest{
		In:       []byte(export),
		Password: "testpassword",
	})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), importResp.KID)

	keyResp, err := service.Key(ctx, &KeyRequest{Key: key.ID().String()})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), keyResp.Key.ID)

	// Check key
	kr := keyring.New(service.vault)
	out, err := kr.EdX25519Key(key.ID())
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID(), key.ID())

	sks, err := kr.EdX25519Keys()
	require.NoError(t, err)
	require.Equal(t, 1, len(sks))

	// Import (bob, ID)
	importResp, err = service.KeyImport(ctx, &KeyImportRequest{
		In: []byte(bob.ID().String()),
	})
	require.NoError(t, err)
	require.Equal(t, bob.ID().String(), importResp.KID)

	// Import (charlie, ID with whitespace)
	importResp, err = service.KeyImport(ctx, &KeyImportRequest{
		In: []byte(charlie.ID().String() + "\n  "),
	})
	require.NoError(t, err)
	require.Equal(t, charlie.ID().String(), importResp.KID)

	// Import (error)
	_, err = service.KeyImport(ctx, &KeyImportRequest{In: []byte{}})
	require.EqualError(t, err, "failed to decode key")
}

func TestKeyImportSaltpack(t *testing.T) {
	msg := `BEGIN EDX25519 KEY MESSAGE.
	9tyMV66eX002JQT sWFyRoiUzCV1DFS Fl2nbyGGteXmU9M XoQcx1V9CKdUCPM
	EoszEpADNLrqULM 2MAcI8XOXSIsAFk 5peBObhA0I9IAZS OOkLndOHMOGHGCd
	dtMkQg08U1C4RtH PMpMj1RyNz9CyBF dNS9qrctSt0r.
	END EDX25519 KEY MESSAGE.`

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	importResp, err := service.KeyImport(ctx, &KeyImportRequest{
		In:       []byte(msg),
		Password: "",
	})
	require.NoError(t, err)
	require.Equal(t, "kex16v9uk4t5wykkklpkrcane3p267n8eu95y3fd55yv4h45m6ku3hyqx2a5fn", importResp.KID)

	keysResp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(keysResp.Keys))
	require.Equal(t, "kex16v9uk4t5wykkklpkrcane3p267n8eu95y3fd55yv4h45m6ku3hyqx2a5fn", keysResp.Keys[0].ID)
}
