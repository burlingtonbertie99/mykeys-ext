package service

import (
	"context"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/vault"
	"github.com/pkg/errors"
)

func setupPassword(vlt *vault.Vault, password string) error {
	salt, err := vlt.Salt()
	if err != nil {
		return err
	}
	key, err := keys.KeyForPassword(password, salt)
	if err != nil {
		return err
	}
	provision := vault.NewProvision(vault.PasswordAuth)
	if err := vlt.Setup(key, provision); err != nil {
		return err
	}
	return nil
}

func unlockPassword(vlt *vault.Vault, password string) (*vault.Provision, error) {
	salt, err := vlt.Salt()
	if err != nil {
		return nil, err
	}
	key, err := keys.KeyForPassword(password, salt)
	if err != nil {
		return nil, err
	}
	return vlt.Unlock(key)
}

func provisionPassword(vlt *vault.Vault, password string) (*vault.Provision, error) {
	salt, err := vlt.Salt()
	if err != nil {
		return nil, err
	}
	key, err := keys.KeyForPassword(password, salt)
	if err != nil {
		return nil, err
	}
	provision := vault.NewProvision(vault.PasswordAuth)
	if err := vlt.Provision(key, provision); err != nil {
		return nil, err
	}

	logger.Infof("Provision (password): %s", provision.ID)
	return provision, nil
}

// AuthPasswordChange (RPC) ...
func (s *service) AuthPasswordChange(ctx context.Context, req *AuthPasswordChangeRequest) (*AuthPasswordChangeResponse, error) {
	old, err := unlockPassword(s.vault, req.Old)
	if err != nil {
		if errors.Cause(err) == vault.ErrInvalidAuth {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}

	if _, err := provisionPassword(s.vault, req.New); err != nil {
		return nil, err
	}

	ok, err := s.vault.Deprovision(old.ID, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to deprovision old password (new password was provisioned)")
	}
	if !ok {
		return nil, errors.Errorf("failed to deprovision, password not found (new password was provisioned)")
	}
	return &AuthPasswordChangeResponse{}, nil
}
