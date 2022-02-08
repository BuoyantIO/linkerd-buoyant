package auth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/credentials"
)

// TokenPerRPCCreds satisfies the credentials.PerRPCCredentials interface
// attaching an up to date token to each RPC request in the GRPC pipeline
// When the token is expired a new token is requested by using the client
// credentials flow
type TokenPerRPCCreds struct {
	secure bool
	ts     oauth2.TokenSource
}

func NewTokenPerRPCCreds(clientID, clientSecret, apiAddr, agentID string, secure bool) credentials.PerRPCCredentials {
	tokenAddrScheme := "http"
	if secure {
		tokenAddrScheme = "https"
	}

	authConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s://%s/agent-token?agent_id=%s", tokenAddrScheme, apiAddr, agentID),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return &TokenPerRPCCreds{
		secure: secure,
		ts:     authConfig.TokenSource(context.Background()),
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
