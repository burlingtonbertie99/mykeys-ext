package rpc_test

import (
	"context"
	"net"
	"testing"

	"github.com/burlingtonbertie99/mykeys-ext/auth/fido2"
	"github.com/burlingtonbertie99/mykeys-ext/auth/rpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func testServer(t *testing.T, addr string) (*grpc.Server, func()) {
	grpcServer := grpc.NewServer()
	fido2.RegisterFIDO2Server(grpcServer, rpc.NewFIDO2Server())

	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	go func() {
		err := grpcServer.Serve(lis)
		require.NoError(t, err)
	}()
	return grpcServer, func() {
		grpcServer.Stop()
	}
}

func testClient(t *testing.T, addr string) (fido2.FIDO2Client, func()) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	client := fido2.NewFIDO2Client(conn)
	return client, func() {
		conn.Close()
	}
}

func TestClientDeviceInfo(t *testing.T) {
	addr := "127.0.0.1:27999"
	_, serverCloseFn := testServer(t, addr)
	defer serverCloseFn()

	client, clientCloseFn := testClient(t, addr)
	defer clientCloseFn()

	ctx := context.TODO()
	resp, err := client.Devices(ctx, &fido2.DevicesRequest{})
	require.NoError(t, err)
	t.Logf("Devices: %+v", resp)

	for _, device := range resp.Devices {
		require.NotEmpty(t, device.Path)

		typeResp, err := client.DeviceType(ctx, &fido2.DeviceTypeRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		if typeResp.Type != fido2.FIDO2Device {
			continue
		}

		infoResp, err := client.DeviceInfo(ctx, &fido2.DeviceInfoRequest{
			Device: device.Path,
		})
		require.NoError(t, err)
		t.Logf("Info: %+v", infoResp.Info)
	}
}
