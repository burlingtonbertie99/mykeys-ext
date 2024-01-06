package service

import (
	"context"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys/encoding"
)

// Rand (RPC) ...
func (s *service) Rand(ctx context.Context, req *RandRequest) (*RandResponse, error) {
	b := keys.RandBytes(int(req.NumBytes))

	enc, err := encodingFromRPC(req.Encoding)
	if err != nil {
		return nil, err
	}

	opts := []encoding.EncodeOption{}
	if req.NoPadding {
		opts = append(opts, encoding.NoPadding())
	}
	if req.Lowercase {
		opts = append(opts, encoding.Lowercase())
	}

	out, err := encoding.Encode(b, enc, opts...)
	if err != nil {
		return nil, err
	}

	return &RandResponse{
		Data: out,
	}, nil
}

func (s *service) RandPassword(ctx context.Context, req *RandPasswordRequest) (*RandPasswordResponse, error) {
	password := keys.RandPassword(int(req.Length))
	return &RandPasswordResponse{
		Password: password,
	}, nil
}
