package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault-client-go"

	"github.com/efortin/kubectl-auth-vault/internal/jwt"
)

type Client struct {
	client *vault.Client
}

type TokenFetcher interface {
	GetOIDCToken(ctx context.Context, path string) (token string, exp int64, err error)
}

func NewClient(address string) (*Client, error) {
	client, err := vault.New(
		vault.WithAddress(address),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) GetOIDCToken(ctx context.Context, path string) (string, int64, error) {
	resp, err := c.client.Read(ctx, path)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read from vault path %s: %w", path, err)
	}

	if resp == nil || resp.Data == nil {
		return "", 0, fmt.Errorf("no data returned from vault path: %s", path)
	}

	tokenRaw, ok := resp.Data["token"]
	if !ok {
		return "", 0, fmt.Errorf("no 'token' field in vault response")
	}

	token, ok := tokenRaw.(string)
	if !ok {
		return "", 0, fmt.Errorf("token is not a string")
	}

	exp, err := jwt.ExtractExp(token)
	if err != nil {
		exp = time.Now().Add(time.Hour).Unix()
	}

	return token, exp, nil
}
