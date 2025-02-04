package vault_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/http/client"
	"github.com/burlingtonbertie99/mykeys-ext/http/server"
	"github.com/burlingtonbertie99/mykeys-ext/vault"
	"github.com/burlingtonbertie99/mykeys/dstore"
	"github.com/burlingtonbertie99/mykeys/encoding"
	"github.com/burlingtonbertie99/mykeys/http"
	"github.com/burlingtonbertie99/mykeys/tsutil"
	"github.com/stretchr/testify/require"
)

func NewTestVaultKey(t *testing.T, clock tsutil.Clock) (*[32]byte, *vault.Provision) {
	key := keys.Bytes32(bytes.Repeat([]byte{0xFF}, 32))
	id := encoding.MustEncode(bytes.Repeat([]byte{0xFE}, 32), encoding.Base62)
	provision := &vault.Provision{
		ID:        id,
		Type:      vault.UnknownAuth,
		CreatedAt: clock.Now(),
	}
	return key, provision
}

type StoreType string

const (
	DB  StoreType = "db"
	Mem StoreType = "mem"
)

type TestVaultOptions struct {
	Unlock bool
	Type   StoreType
	Clock  tsutil.Clock
}

func NewTestVault(t *testing.T, opts *TestVaultOptions) (*vault.Vault, func()) {
	if opts == nil {
		opts = &TestVaultOptions{}
	}
	if opts.Type == "" {
		opts.Type = Mem
	}
	if opts.Clock == nil {
		opts.Clock = tsutil.NewTestClock()
	}

	var st vault.Store
	var closeFn func()
	switch opts.Type {
	case Mem:
		st, closeFn = newTestMem(t)
	case DB:
		st, closeFn = newTestDB(t)
	}

	vlt := vault.New(st, vault.WithClock(opts.Clock))

	if opts.Unlock {
		key, provision := NewTestVaultKey(t, opts.Clock)
		err := vlt.Setup(key, provision)
		require.NoError(t, err)
		_, err = vlt.Unlock(key)
		require.NoError(t, err)
	}
	return vlt, closeFn
}

func newTestMem(t *testing.T) (vault.Store, func()) {
	mem := vault.NewMem()
	err := mem.Open()
	require.NoError(t, err)
	closeFn := func() {
		mem.Close()
	}
	return mem, closeFn
}

func newTestDB(t *testing.T) (vault.Store, func()) {
	path := testPath()
	db := vault.NewDB(path)
	err := db.Open()
	require.NoError(t, err)
	close := func() {
		err := db.Close()
		require.NoError(t, err)
		_ = os.RemoveAll(path)
	}
	return db, close
}

// func testSeed(b byte) *[32]byte {
// 	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
// }

type testEnv struct {
	clock      tsutil.Clock
	httpServer *httptest.Server
	srv        *server.Server
	closeFn    func()
}

func newTestEnv(t *testing.T, logger server.Logger) *testEnv {
	if logger == nil {
		logger = client.NewLogger(client.LogLevel(-1))
	}
	clock := tsutil.NewTestClock()
	fi := dstore.NewMem()
	fi.SetClock(clock)
	client := http.NewClient()

	rds := server.NewRedisTest(clock)
	srv := server.New(fi, rds, client, clock, logger)
	srv.SetClock(clock)
	srv.SetInternalAuth("testtoken")
	_ = srv.SetInternalKey("6a169a699f7683c04d127504a12ace3b326e8b56a61a9b315cf6b42e20d6a44a")
	handler := server.NewHandler(srv)
	httpServer := httptest.NewServer(handler)
	srv.URL = httpServer.URL

	return &testEnv{clock, httpServer, srv, func() { httpServer.Close() }}
}

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.vdb", keys.RandFileName()))
}

func newTestClient(t *testing.T, env *testEnv) *vault.Client {
	cl, err := client.New(env.httpServer.URL)
	require.NoError(t, err)
	cl.SetHTTPClient(env.httpServer.Client())
	cl.SetClock(env.clock)
	return vault.NewClient(cl)
}

func TestIsEmpty(t *testing.T) {
	db, closeFn := newTestDB(t)
	defer closeFn()
	vlt := vault.New(db)
	empty, err := vlt.IsEmpty()
	require.NoError(t, err)
	require.True(t, empty)
}

func TestErrors(t *testing.T) {
	// vault.SetLogger(vault.NewLogger(vault.DebugLevel))
	var err error
	env := newTestEnv(t, nil) // vault.NewLogger(vault.DebugLevel))
	defer env.closeFn()

	vlt, closeFn := NewTestVault(t, nil)
	defer closeFn()

	err = vlt.Set(vault.NewItem("key1", []byte("mysecretdata"), "", time.Now()))
	require.EqualError(t, err, "vault is locked")

	key := keys.Rand32()
	err = vlt.Setup(key, vault.NewProvision(vault.UnknownAuth))
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("mysecretdata"), "", time.Now()))
	require.NoError(t, err)
	vlt.Lock()

	_, err = vlt.Get("key1")
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")
	_, err = vlt.ItemHistory("key1")
	require.EqualError(t, err, "vault is locked")
}

func TestVaultGet(t *testing.T) {
	db, closeFn := newTestDB(t)
	defer closeFn()
	vlt := vault.New(db)
	testVaultGet(t, vlt)
}

func testVaultGet(t *testing.T, vlt *vault.Vault) {
	var err error
	key := keys.Rand32()
	provision := vault.NewProvision(vault.UnknownAuth)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 0, len(items))

	out, err := vlt.Get("abc")
	require.NoError(t, err)
	require.Nil(t, out)

	_, err = vlt.Get("")
	require.EqualError(t, err, "empty id")

	now := time.Now()

	// Set "abc"
	item := vault.NewItem("abc", []byte("password"), "type1", now)
	err = vlt.Set(item)
	require.NoError(t, err)

	out, err = vlt.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "abc", out.ID)
	require.Equal(t, []byte("password"), out.Data)
	require.Equal(t, tsutil.Millis(now), tsutil.Millis(out.Timestamp))

	// Update
	item.Data = []byte("newpassword")
	err = vlt.Set(item)
	require.NoError(t, err)

	out, err = vlt.Get("abc")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, "abc", out.ID)
	require.Equal(t, []byte("newpassword"), out.Data)
	require.Equal(t, tsutil.Millis(now), tsutil.Millis(out.Timestamp))

	// Set "xyz"
	err = vlt.Set(vault.NewItem("xyz", []byte("xpassword"), "type2", time.Now()))
	require.NoError(t, err)

	items, err = vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, items[0].ID, "abc")
	require.Equal(t, items[1].ID, "xyz")

	// Delete
	ok, err := vlt.Delete("abc")
	require.NoError(t, err)
	require.True(t, ok)

	item3, err := vlt.Get("abc")
	require.NoError(t, err)
	require.Nil(t, item3)

	ok2, err := vlt.Delete("abc")
	require.NoError(t, err)
	require.False(t, ok2)

	items, err = vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, items[0].ID, "xyz")
}

func TestSetupUnlockProvision(t *testing.T) {
	db, closeFn := newTestDB(t)
	defer closeFn()
	testSetupUnlockProvision(t, db)
}

func testSetupUnlockProvision(t *testing.T, st vault.Store) {
	var err error

	vlt := vault.New(st)

	clock := tsutil.NewTestClock()
	key, provision := NewTestVaultKey(t, clock)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.NoError(t, err)

	vlt.Lock()

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Get("key1")
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	err = vlt.Set(vault.NewItem("key1", []byte("password"), "", time.Now()))
	require.NoError(t, err)

	vlt.Lock()

	_, err = vlt.Items()
	require.EqualError(t, err, "vault is locked")

	_, err = vlt.Delete("key1")
	require.EqualError(t, err, "vault is locked")

	key2 := keys.Bytes32(bytes.Repeat([]byte{0x02}, 32))
	_, err = vlt.Unlock(key2)
	require.EqualError(t, err, "invalid auth")

	// Unlock
	_, err = vlt.Unlock(key)
	require.NoError(t, err)
	provision2 := vault.NewProvision(vault.UnknownAuth)
	key3 := keys.Rand32()
	err = vlt.Provision(key3, provision2)
	require.NoError(t, err)

	// Deprovision
	ok, err := vlt.Deprovision(provision.ID, false)
	require.NoError(t, err)
	require.True(t, ok)

	paths, err := vaultPaths(vlt, "/provision")
	require.NoError(t, err)
	require.Equal(t, []string{"/provision/" + provision2.ID}, paths)

	// // Don't deprovision last
	_, err = vlt.Deprovision(provision2.ID, false)
	require.EqualError(t, err, "failed to deprovision: last auth")

	ok, err = vlt.Deprovision(provision2.ID, true)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSetErrors(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	err = vlt.Set(vault.NewItem("", nil, "", time.Time{}))
	require.EqualError(t, err, "invalid id")
}

func TestProtocol(t *testing.T) {
	st, closeFn := newTestDB(t)
	defer closeFn()
	vlt := vault.New(st)

	var err error

	// Setup
	salt := bytes.Repeat([]byte{0x01}, 32)
	key, err := keys.KeyForPassword("password123", salt)
	require.NoError(t, err)
	provision := &vault.Provision{
		ID:        encoding.MustEncode(bytes.Repeat([]byte{0x02}, 32), encoding.Base62),
		Type:      vault.UnknownAuth,
		CreatedAt: time.Now(),
	}
	err = vlt.Setup(key, provision)
	require.NoError(t, err)
	_, err = vlt.Unlock(key)
	require.NoError(t, err)

	// Create item
	item := vault.NewItem("testid1", []byte("testpassword"), "", time.Now())
	err = vlt.Set(item)
	require.NoError(t, err)

	paths, err := vaultPaths(vlt, "")
	require.NoError(t, err)
	require.Equal(t, []string{
		"/auth/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/item/testid1",
		"/provision/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000001/auth/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000002/provision/0TWD4V5tkyUQGc5qXvlBDd2Fj97aqsMoBGJJjsttG4I",
		"/push/000000000000003/item/testid1",
		"/sync/push",
		"/sync/rsalt",
	}, paths)

	paths, err = vaultPaths(vlt, "/auth")
	require.NoError(t, err)
	require.Equal(t, []string{"/auth/" + provision.ID}, paths)

	items, err := vlt.Items()
	require.NoError(t, err)
	require.Equal(t, 1, len(items))
	require.Equal(t, "testid1", items[0].ID)
}

func vaultPaths(vlt *vault.Vault, prefix string) ([]string, error) {
	docs, err := vlt.Store().List(&vault.ListOptions{Prefix: prefix})
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, doc := range docs {
		paths = append(paths, doc.Path)

	}
	return paths, nil
}

// TODO: Create test db in testdata for backward compatibility testing.
// func TestVaultDB(t *testing.T) {
// 	db := vault.NewDB(filepath.Join("testdata", "vault.vdb"))
// 	err := db.Open()
// 	require.NoError(t, err)
// 	defer db.Close()

// 	vlt := vault.New(db)

// 	err = vlt.UnlockWithPassword("password", false)
// 	require.NoError(t, err)

// 	ks, err := vlt.Keys()
// 	require.NoError(t, err)
// 	for _, key := range ks {
// 		_, err := key.AsEdX25519Public()
// 		require.NoError(t, err)
// 	}
// 	require.Equal(t, 20, len(ks))

// 	pks, err := vlt.EdX25519PublicKeys()
// 	require.Equal(t, 19, len(pks))
// }
