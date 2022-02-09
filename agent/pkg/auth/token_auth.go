package auth

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials"
)

// TokenPerRPCCreds satisfies the credentials.PerRPCCredentials interface
// attaching an up to date token to each RPC request in the GRPC pipeline.
// When the token is expired a new token is requested by using the client
// credentials flow.
type TokenPerRPCCreds struct {
	secure bool
	ts     oauth2.TokenSource
}

func NewTokenPerRPCCreds(ts oauth2.TokenSource, secure bool) credentials.PerRPCCredentials {
	return &TokenPerRPCCreds{
		ts:     ts,
		secure: secure,
	}
}

func (t *TokenPerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := t.ts.Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (t *TokenPerRPCCreds) RequireTransportSecurity() bool {
	return t.secure
}
