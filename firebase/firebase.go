package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// Client represents firebase client.
type Client struct {
	auth *auth.Client
}

// VerifyToken verifies the signature of the provided token and returns token claims.
func (c *Client) VerifyToken(ctx context.Context, token string) (claims map[string]interface{}, err error) {
	t, err := c.auth.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return t.Claims, nil
}

// NewClient creates new instance of firebase client.
func NewClient(credentialFile string) (*Client, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase: could not create new app: %s", err)
	}
	ac, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: could not get auth: %s", err)
	}

	return &Client{auth: ac}, nil
}
