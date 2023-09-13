package fakeauthenticator

import (
	"context"
)

// Client represents firebase client.
type Client struct {
	tmp map[string]interface{}
}

// VerifyToken verifies the signature of the provided token and returns token claims.
func (c *Client) VerifyToken(ctx context.Context, token string) (claims map[string]interface{}, err error) {
	return c.tmp, nil
}

// NewClient creates new instance of firebase client.
func NewClient(credentialFile string) (*Client, error) {
	fakeClaims := map[string]interface{}{
		"user_id": "fake_temporary_id",
	}
	return &Client{fakeClaims}, nil
}
