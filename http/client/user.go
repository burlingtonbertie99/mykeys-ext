package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/burlingtonbertie99/mykeys-ext/http/api"
	"github.com/pkg/errors"
)

// UserSearch ...
func (c *Client) UserSearch(ctx context.Context, query string, limit int) (*api.UserSearchResponse, error) {
	params := url.Values{}
	params.Add("q", query)
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	resp, err := c.Request(ctx, &Request{Method: "GET", Path: "/user/search", Params: params})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.Errorf("/user/search not found")
	}

	// TODO: Support paging
	var out api.UserSearchResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// User ...
func (c *Client) User(ctx context.Context, kid keys.ID) (*api.UserResponse, error) {
	params := url.Values{}
	resp, err := c.Request(ctx, &Request{Method: "GET", Path: "/user/" + kid.String(), Params: params})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.UserResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
