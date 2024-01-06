package service

import (
	"context"
	"fmt"
	"github.com/burlingtonbertie99/mykeys/http"
	"time"
)

// AdminSignURL ...
func (s *service) AdminSignURL(ctx context.Context, req *AdminSignURLRequest) (*AdminSignURLResponse, error) {
	kid, err := s.lookup(ctx, req.Signer, nil)
	if err != nil {
		return nil, err
	}
	key, err := s.edx25519Key(kid)
	if err != nil {
		return nil, err
	}

	auth, err := http.NewAuth(req.Method, req.URL, "", time.Now(), key)
	if err != nil {
		return nil, err
	}

	curl := fmt.Sprintf("curl -X %s -H \"Authorization: %s\" %s", req.Method, auth.Header(), auth.URL.String())

	return &AdminSignURLResponse{
		Auth: auth.Header(),
		URL:  auth.URL.String(),
		CURL: curl,
	}, nil
}

func (s *service) AdminCheck(ctx context.Context, req *AdminCheckRequest) (*AdminCheckResponse, error) {
	kid, err := s.lookup(ctx, req.Signer, nil)
	if err != nil {
		return nil, err
	}
	key, err := s.edx25519Key(kid)
	if err != nil {
		return nil, err
	}

	if err := s.client.AdminCheck(ctx, req.Check, key); err != nil {
		return nil, err
	}

	return &AdminCheckResponse{}, nil
}
