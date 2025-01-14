package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/burlingtonbertie99/mykeys-ext/auth/fido2"
	"github.com/pkg/errors"
)

func fido2AuthSetup(ctx context.Context, client *Client, clientName string, device string, pin string) (string, error) {
	if device == "" {
		d, err := selectDevice(ctx, client)
		if err != nil {
			return "", err
		}
		device = d
	}

	if pin == "" {
		p, err := readPassword("Enter your PIN:", true)
		if err != nil {
			return "", err
		}
		pin = p
	}

	fmt.Fprintln(os.Stderr, "Let's create a credential, you may need to interact with the key...")
	if _, err := client.RPCClient().AuthSetup(ctx, &AuthSetupRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Device: device,
	}); err != nil {
		return "", err
	}

	fmt.Fprintln(os.Stderr, "Getting the credential, you may need to interact with the key...")
	unlockResp, err := client.RPCClient().AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}

	return unlockResp.AuthToken, nil
}

func fido2AuthUnlock(ctx context.Context, client *Client, clientName string, pin string) (string, error) {
	if pin == "" {
		p, err := readPassword("Enter your PIN:", true)
		if err != nil {
			return "", err
		}
		pin = p
	}

	fmt.Fprintln(os.Stderr, "Getting the credential, you may need to interact with the key...")
	unlock, err := client.RPCClient().AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}
	return unlock.AuthToken, nil
}

func fido2AuthProvision(ctx context.Context, client *Client, device string, pin string) error {
	if device == "" {
		d, err := selectDevice(ctx, client)
		if err != nil {
			return err
		}
		device = d
	}

	if pin == "" {
		p, err := readPassword("Enter your PIN:", true)
		if err != nil {
			return err
		}
		pin = p
	}

	// Generate
	if err := fido2AuthGenerate(ctx, client, pin, device); err != nil {
		return err
	}
	// Credential
	if err := fido2AuthCredential(ctx, client, pin, device); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Successfully provisioned (FIDO2 HMAC-Secret).")
	return nil
}

func fido2AuthGenerate(ctx context.Context, client *Client, pin string, device string) error {
	fmt.Fprintln(os.Stderr, "Let's create a credential, you may need to interact with the key...")
	if _, err := client.RPCClient().AuthProvision(ctx, &AuthProvisionRequest{
		Device:   device,
		Secret:   pin,
		Type:     FIDO2HMACSecretAuth,
		Generate: true,
	}); err != nil {
		return err
	}
	return nil
}

func fido2AuthCredential(ctx context.Context, client *Client, pin string, device string) error {
	fmt.Fprintln(os.Stderr, "Getting the credential, you may need to interact with the key (again)...")
	if _, err := client.RPCClient().AuthProvision(ctx, &AuthProvisionRequest{
		Device: device,
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
	}); err != nil {
		return err
	}
	return nil
}

func selectDevice(ctx context.Context, client *Client) (string, error) {
	devicesResp, err := client.FIDO2Client().Devices(ctx, &fido2.DevicesRequest{})
	if err != nil {
		return "", err
	}
	if len(devicesResp.Devices) == 0 {
		return "", errors.Errorf("no devices found")
	}
	if len(devicesResp.Devices) == 1 {
		return devicesResp.Devices[0].Path, nil
	}
	fmt.Fprintf(os.Stderr, "Choose a device:\n")
	for i, d := range devicesResp.Devices {
		fmt.Fprintf(os.Stderr, "(%d) %s (%s)\n", i+1, d.Product, d.Manufacturer)
	}
	reader := bufio.NewReader(os.Stdin)
	read, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	n, err := strconv.Atoi(strings.TrimRight(read, " \n"))
	if err != nil {
		return "", errors.Errorf("invalid input: %s", read)
	}
	if n < 1 || n > len(devicesResp.Devices) {
		return "", errors.Errorf("invalid input (%d)", n)
	}
	return devicesResp.Devices[n-1].Path, nil
}
