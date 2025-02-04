package client

import (
	"context"
	"net/url"

	"github.com/burlingtonbertie99/mykeys"
)

// AdminCheck performs user & sigchain associated with key by an admin.
// The server periodically checks users and sigchains, but this tells the server
// to do it right away.
// If check is a kid, will check that key.
// If check is "all" (not recommended), it will check all keys.
func (c *Client) AdminCheck(ctx context.Context, check string, admin *keys.EdX25519Key) error {
	params := url.Values{}
	_, err := c.Request(ctx, &Request{Method: "POST", Path: "/admin/check/" + check, Params: params, Key: admin})
	if err != nil {
		return err
	}
	return nil
}
