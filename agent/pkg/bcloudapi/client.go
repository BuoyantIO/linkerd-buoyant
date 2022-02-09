package bcloudapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	agentTokenEndpoint   = "/agent-token"
	tokenEndpoint        = "/token"
	regiserAgentEndpoint = "/register-agent"
)

// AgentInfo contains all data to describe an agent that has been
// registered in Bcloud.
type AgentInfo struct {
	AgentName  string `json:"agent_name"`
	AgentID    string `json:"agent_id"`
	IsNewAgent bool   `json:"is_new_agent"`
}

type Client interface {
	RegisterAgent(ctx context.Context, agentName string) (*AgentInfo, error)
	AgentTokenSource(ctx context.Context, agentID string) oauth2.TokenSource
}

// ApiClient is a client that talks to the Bcloud client API
type ApiClient struct {
	clientID        string
	clientSecret    string
	tokenAuthConfig *clientcredentials.Config
	addrScheme      string
	apiAddr         string
}

// New creates a new ApiClient
func New(clientID, clientSecret, apiAddr string, secure bool) *ApiClient {
	addrScheme := "http"
	if secure {
		addrScheme = "https"
	}

	tokenAuthConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s://%s%s", addrScheme, apiAddr, tokenEndpoint),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return &ApiClient{
		clientID:        clientID,
		clientSecret:    clientSecret,
		tokenAuthConfig: tokenAuthConfig,
		addrScheme:      addrScheme,
		apiAddr:         apiAddr,
	}
}

// RegisterAgent registers the agent with Bcloud
func (c *ApiClient) RegisterAgent(ctx context.Context, agentName string) (*AgentInfo, error) {
	client := c.tokenAuthConfig.Client(ctx)
	url := fmt.Sprintf("%s://%s%s?%s=%s", c.addrScheme, c.apiAddr, regiserAgentEndpoint, k8s.AgentNameKey, agentName)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent registration api returned: %d", rsp.StatusCode)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	info := &AgentInfo{}
	if err := json.Unmarshal(data, info); err != nil {
		return nil, err
	}

	return info, nil
}

// AgentTokenSource returns a token source for a particular agent
func (c *ApiClient) AgentTokenSource(ctx context.Context, agentID string) oauth2.TokenSource {
	authConfig := &clientcredentials.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		TokenURL:     fmt.Sprintf("%s://%s%s?agent_id=%s", c.addrScheme, c.apiAddr, agentTokenEndpoint, agentID),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return authConfig.TokenSource(ctx)
}
