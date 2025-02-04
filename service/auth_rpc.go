package service

import (
	"context"
	"encoding/json"
	"os"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/vault"
	"github.com/burlingtonbertie99/mykeys/encoding"
	"github.com/burlingtonbertie99/mykeys/tsutil"
	"github.com/pkg/errors"
)

// AuthSetup (RPC) ...
func (s *service) AuthSetup(ctx context.Context, req *AuthSetupRequest) (*AuthSetupResponse, error) {
	logger.Infof("Auth setup...")
	status, err := s.vault.Status()
	if err != nil {
		return nil, err
	}
	if status != vault.SetupNeeded {
		return nil, errors.Errorf("auth already setup")
	}

	if err := s.auth.setup(ctx, s.vault, req); err != nil {
		return nil, err
	}

	// If setting up auth, and service database exists we should nuke it since the
	// pre-existing key is different. The database will be rebuilt on Open.
	path, err := s.env.AppPath(cdbPath, false)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Checking for existing db...")
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		logger.Debugf("Removing existing db: %s", path)
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	return &AuthSetupResponse{}, nil
}

// AuthVault (RPC) ...
func (s *service) AuthVault(ctx context.Context, req *AuthVaultRequest) (*AuthVaultResponse, error) {
	logger.Infof("Auth vault...")
	status, err := s.vault.Status()
	if err != nil {
		return nil, err
	}
	if status != vault.SetupNeeded {
		return nil, errors.Errorf("auth already setup")
	}

	otkSeed, err := encoding.Decode(req.Phrase, encoding.BIP39)
	if err != nil {
		return nil, err
	}
	if len(otkSeed) != 32 {
		return nil, errors.Errorf("invalid byte length for otk")
	}
	otk := keys.NewEdX25519KeyFromSeed(keys.Bytes32(otkSeed))

	remoteBytes, err := s.client.ShareOpen(ctx, otk)
	if err != nil {
		return nil, err
	}
	if remoteBytes == nil {
		return nil, errors.Errorf("vault not found")
	}
	var remote vault.Remote
	if err := json.Unmarshal(remoteBytes, &remote); err != nil {
		return nil, errors.Wrapf(err, "invalid vault remote bytes")
	}
	if err := s.vault.Clone(ctx, &remote); err != nil {
		return nil, err
	}

	return &AuthVaultResponse{}, nil
}

// AuthUnlock (RPC) ...
func (s *service) AuthUnlock(ctx context.Context, req *AuthUnlockRequest) (*AuthUnlockResponse, error) {
	token, err := s.unlock(ctx, req)
	if err != nil {
		// if errors.Cause(err).Error() == "pin required" {
		// 	return nil, status.Error(codes.FailedPrecondition, "pin required")
		// }
		return nil, err
	}
	return &AuthUnlockResponse{
		AuthToken: token,
	}, nil
}

// AuthLock (RPC) ...
func (s *service) AuthLock(ctx context.Context, req *AuthLockRequest) (*AuthLockResponse, error) {
	s.lock()
	return &AuthLockResponse{}, nil
}

// AuthProvision (RPC) ...
func (s *service) AuthProvision(ctx context.Context, req *AuthProvisionRequest) (*AuthProvisionResponse, error) {
	provision, err := s.auth.provision(ctx, s.vault, req)
	if err != nil {
		return nil, err
	}
	return &AuthProvisionResponse{
		Provision: provisionToRPC(provision),
	}, nil
}

// AuthDeprovision (RPC) ...
func (s *service) AuthDeprovision(ctx context.Context, req *AuthDeprovisionRequest) (*AuthDeprovisionResponse, error) {
	ok, err := s.vault.Deprovision(req.ID, false)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(req.ID)
	}

	// TODO: If FIDO2 resident key and supports credMgmt remove from the device also?

	return &AuthDeprovisionResponse{}, nil
}

// AuthProvisions (RPC) ...
func (s *service) AuthProvisions(ctx context.Context, req *AuthProvisionsRequest) (*AuthProvisionsResponse, error) {
	provisions, err := s.vault.Provisions()
	if err != nil {
		return nil, err
	}

	out := make([]*AuthProvision, 0, len(provisions))
	for _, provision := range provisions {
		out = append(out, provisionToRPC(provision))
	}

	return &AuthProvisionsResponse{
		Provisions: out,
	}, nil
}

type testClientAuth struct {
	token string
}

func newTestClientAuth(token string) testClientAuth {
	return testClientAuth{token: token}
}

func (a testClientAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	if a.token == "" {
		return nil, nil
	}
	return map[string]string{
		"authorization": a.token,
	}, nil
}

func (a testClientAuth) RequireTransportSecurity() bool {
	// For test client
	return false
}

func provisionToRPC(p *vault.Provision) *AuthProvision {
	return &AuthProvision{
		ID:        p.ID,
		Type:      authTypeToRPC(p.Type),
		AAGUID:    p.AAGUID,
		NoPin:     p.NoPin,
		CreatedAt: tsutil.Millis(p.CreatedAt),
	}
}

func authTypeToRPC(t vault.AuthType) AuthType {
	switch t {
	case vault.PasswordAuth:
		return PasswordAuth
	case vault.PaperKeyAuth:
		return PaperKeyAuth
	case vault.FIDO2HMACSecretAuth:
		return FIDO2HMACSecretAuth
	default:
		return UnknownAuth
	}
}

func matchAAGUID(provisions []*vault.Provision, aaguid string) *vault.Provision {
	for _, provision := range provisions {
		if provision.AAGUID == aaguid {
			return provision
		}
	}
	return nil
}

func (s *service) AuthRecover(ctx context.Context, req *AuthRecoverRequest) (*AuthRecoverResponse, error) {
	if req.PaperKey == "" {
		return nil, errors.Errorf("no paper key specified")
	}
	if req.NewPassword == "" {
		return nil, errors.Errorf("no password specified")
	}
	unlockResp, err := s.AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: req.PaperKey,
		Type:   PaperKeyAuth,
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.AuthProvision(ctx, &AuthProvisionRequest{
		Secret: req.NewPassword,
		Type:   PasswordAuth,
	}); err != nil {
		return nil, err
	}

	return &AuthRecoverResponse{
		AuthToken: unlockResp.AuthToken,
	}, nil
}

func (s *service) AuthReset(ctx context.Context, req *AuthResetRequest) (*AuthResetResponse, error) {
	if s.unlocked {
		return nil, errors.Wrapf(errors.Errorf("auth is unlocked"), "failed to reset")
	}

	if req.AppName != s.env.AppName() {
		return nil, errors.Wrapf(errors.Errorf("invalid app name"), "failed to reset")
	}

	if err := s.vault.Reset(); err != nil {
		return nil, err
	}

	path, err := s.env.AppPath(cdbPath, false)
	if err != nil {
		return nil, err
	}
	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}

	return &AuthResetResponse{}, nil
}
