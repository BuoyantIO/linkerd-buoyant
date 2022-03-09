package bcloudapi

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// tokenPerRPCCreds satisfies the credentials.PerRPCCredentials interface
// attaching an up to date token to each RPC request in the GRPC pipeline.
// When the token is expired a new token is requested by using the client
// credentials flow.
type tokenPerRPCCreds struct {
	noTLS bool
	ts    oauth.TokenSource
}

func newTokenPerRPCCreds(ts oauth2.TokenSource, noTLS bool) credentials.PerRPCCredentials {
	return &tokenPerRPCCreds{
		ts:    oauth.TokenSource{TokenSource: ts},
		noTLS: noTLS,
	}
}

func (t *tokenPerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if t.noTLS {
		token, err := t.ts.Token()
		if err != nil {
			return nil, err
		}

		return map[string]string{
			"authorization": token.Type() + " " + token.AccessToken,
		}, nil
	}

	return t.ts.GetRequestMetadata(ctx, uri...)
}

func (t *tokenPerRPCCreds) RequireTransportSecurity() bool {
	return !t.noTLS
}
