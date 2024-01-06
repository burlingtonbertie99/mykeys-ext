package client_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/http/client"
	"github.com/burlingtonbertie99/mykeys-ext/http/server"
	"github.com/stretchr/testify/require"
)

func TestDisco(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	// logger = NewLogger(DebugLevel)

	env, closeFn := newEnv(t, server.NoLevel)
	defer closeFn()

	aliceClient := newTestClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	bobClient := newTestClient(t, env)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// Put
	err := aliceClient.DiscoSave(context.TODO(), alice, bob.ID(), client.Offer, "hi", time.Minute)
	require.NoError(t, err)

	// Get
	out, err := bobClient.Disco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Equal(t, "hi", out)

	// Get (again)
	out, err = bobClient.Disco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.DiscoSave(context.TODO(), alice, bob.ID(), client.Offer, "hi2", time.Minute)
	require.NoError(t, err)

	// Delete
	err = aliceClient.DiscoDelete(context.TODO(), alice, bob.ID())
	require.NoError(t, err)

	// Get (deleted)
	out, err = bobClient.Disco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)

	// Put
	err = aliceClient.DiscoSave(context.TODO(), alice, bob.ID(), client.Offer, "hi3", time.Millisecond)
	require.NoError(t, err)

	// Get (expired)
	time.Sleep(time.Millisecond)
	out, err = bobClient.Disco(context.TODO(), alice.ID(), bob, client.Offer)
	require.NoError(t, err)
	require.Empty(t, out)
}
