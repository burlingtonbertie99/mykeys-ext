package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys/http"
	"github.com/burlingtonbertie99/mykeys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestSignVerify(t *testing.T) {
	var err error
	env := newTestEnv(t)
	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	// Sign with kid (alice)
	message := "I'm alice"
	signResp, err := aliceService.Sign(context.TODO(), &SignRequest{Data: []byte(message), Signer: alice.ID().String()})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)

	verifyResp, err := aliceService.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)

	testUserSetupGithub(t, env, aliceService, alice, "alice")

	// Sign with alice@github
	signResp, err = aliceService.Sign(context.TODO(), &SignRequest{Data: []byte(message), Signer: "alice@github"})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)

	verifyResp, err = aliceService.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.NotNil(t, verifyResp.Signer)
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)
	require.NotNil(t, verifyResp.Signer.User)
	require.Equal(t, "alice", verifyResp.Signer.User.Name)

	// Bob
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)

	// Verify at bob
	verifyResp, err = bobService.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.NotNil(t, verifyResp.Signer)
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)
	require.Nil(t, verifyResp.Signer.User)

	// View key (do not import)
	_, err = bobService.Key(context.TODO(), &KeyRequest{Key: alice.ID().String(), Update: true})
	require.NoError(t, err)

	// Re-verify
	verifyResp, err = bobService.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.NotNil(t, verifyResp.Signer)
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)
	require.Nil(t, verifyResp.Signer.User)

	// Sign (not found)
	key := keys.GenerateEdX25519Key()
	_, err = aliceService.Sign(context.TODO(), &SignRequest{Data: []byte(message), Signer: key.ID().String()})
	require.EqualError(t, err, fmt.Sprintf("%s not found", key.ID()))
}

func TestSignStream(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	testSignStream(t, env, service, bytes.Repeat([]byte{0x31}, 5), alice.ID().String())
	testSignStream(t, env, service, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String())
	// TODO: Test timeout if data stops streaming
}

func testSignStream(t *testing.T, env *testEnv, service *service, plaintext []byte, signer string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var clientOut bytes.Buffer
	cl, clientCloseFn := newTestRPCClient(t, service, env, "", &clientOut)
	defer clientCloseFn()

	streamClient, streamErr := cl.RPCClient().SignStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&SignInput{
			Signer:  signer,
			Armored: true,
		})
		require.NoError(t, err)
		for chunk := 0; true; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(plaintext) {
				e = len(plaintext)
				done = true
			}
			logger.Debugf("(Test) Send chunk %d", len(plaintext[s:e]))
			err := streamClient.Send(&SignInput{
				Data: plaintext[s:e],
			})
			require.NoError(t, err)
			if done {
				logger.Debugf("(Test) Send done")
				break
			}
		}
		logger.Debugf("(Test) Close send")
		closeErr2 := streamClient.CloseSend()
		require.NoError(t, closeErr2)
	}()

	var buf bytes.Buffer
	for {
		resp, recvErr := streamClient.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				logger.Debugf("(Test) Recv EOF, done")
				break
			}
			require.NoError(t, recvErr)
		}
		logger.Infof("(Test) Recv %d", len(resp.Data))
		_, writeErr := buf.Write(resp.Data)
		require.NoError(t, writeErr)
	}

	// Verify (from Saltpack)
	data := buf.Bytes()
	out, sout, err := saltpack.Verify(data)
	require.NoError(t, err)
	if signer != "" {
		require.Equal(t, sout.String(), signer)
	}
	require.Equal(t, plaintext, out)

	// Verify stream
	outClient, streamErr := cl.RPCClient().VerifyStream(ctx)
	require.NoError(t, streamErr)

	go func() {
		done := false
		for chunk := 0; ; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(data) {
				e = len(data)
				done = true
			}
			err := outClient.Send(&VerifyInput{
				Data: data[s:e],
			})
			require.NoError(t, err)
			if done {
				break
			}
		}
		closeErr2 := outClient.CloseSend()
		require.NoError(t, closeErr2)
	}()

	var bufOut bytes.Buffer
	for {
		resp, recvErr := outClient.Recv()
		require.NoError(t, recvErr)
		if len(resp.Data) == 0 {
			break
		}
		_, writeErr := bufOut.Write(resp.Data)
		require.NoError(t, writeErr)
	}

	require.Equal(t, plaintext, bufOut.Bytes())
}

func TestSignVerifyAttachedFile(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	b := []byte("test message")
	inPath := keys.RandTempPath()
	outPath := inPath + ".signed"
	verifiedPath := inPath + ".ver"

	defer os.Remove(inPath)
	defer os.Remove(outPath)
	defer os.Remove(verifiedPath)

	writeErr := ioutil.WriteFile(inPath, b, 0644)
	require.NoError(t, writeErr)

	var aliceOut bytes.Buffer
	aliceClient, aliceClientCloseFn := newTestRPCClient(t, aliceService, env, "", &aliceOut)
	defer aliceClientCloseFn()

	err := signFile(aliceClient, alice.ID().String(), true, false, inPath, outPath)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("out: %s.signed\n", inPath), aliceOut.String())

	var bobOut bytes.Buffer
	bobClient, bobClientCloseFn := newTestRPCClient(t, bobService, env, "", &bobOut)
	defer bobClientCloseFn()

	_, err = verifyFile(bobClient, outPath, verifiedPath, alice.ID().String())
	require.NoError(t, err)

	bout, err := ioutil.ReadFile(verifiedPath)
	require.NoError(t, err)
	require.Equal(t, b, bout)
	os.Remove(verifiedPath)

	// Test nextPath
	// out, err := verifyFile(bobClient, true, outPath, "", alice.ID().String())
	// require.NoError(t, err)
	// require.Equal(t, inPath+"-1", out)
	// os.Remove(out)
}

func TestVerifyUnverified(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	tub := testUserSetupGithub(t, env, bobService, bob, "bob")

	testPull(t, aliceService, bob.ID())

	env.clock.Add(time.Hour * 24)

	// Set 500 error for bob@github
	env.client.SetProxy(tub.URL, func(ctx context.Context, req *http.Request) http.ProxyResponse {
		return http.ProxyResponse{Err: http.Err{Code: 500}}
	})

	// Sign (bob)
	signResp, err := bobService.Sign(context.TODO(), &SignRequest{
		Signer: bob.ID().String(),
		Data:   []byte("test"),
	})
	require.NoError(t, err)

	// Verify (bob, error)
	_, err = aliceService.Verify(context.TODO(), &VerifyRequest{
		Data: signResp.Data,
	})
	require.EqualError(t, err, "user bob@github has failed status connection-fail")
}
