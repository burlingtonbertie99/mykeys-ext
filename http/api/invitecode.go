package api

import "github.com/burlingtonbertie99/mykeys"

// InviteCodeCreateResponse ...
type InviteCodeCreateResponse struct {
	Code string `json:"code"`
}

// InviteCodeResponse ...
type InviteCodeResponse struct {
	Sender    keys.ID `json:"sender"`
	Recipient keys.ID `json:"recipient"`
}
