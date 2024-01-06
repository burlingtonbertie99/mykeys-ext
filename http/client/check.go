package client

import (
	"context"
	"net/url"

	"github.com/burlingtonbertie99/mykeys"
)

// Check user & sigchain associated with edx25519 key.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
func (c *Client) Check(ctx context.Context, key *keys.EdX25519Key) error {
	params := url.Values{}
	if _, err := c.Request(ctx, &Request{Method: "POST", Path: "/check", Params: params, Key: key}); err != nil {
		return err
	}
	return nil
}
