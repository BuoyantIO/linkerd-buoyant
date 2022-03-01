package bcloudapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/credentials"
)

const (
	agentTokenEndpoint    = "/agent-token"
	tokenEndpoint         = "/token"
	registerAgentEndpoint = "/register-agent"
	agentManifestEndpoint = "/agent.yaml"
)

type jsonError struct {
	Error string `json:"error"`
}

// AgentIdentifier identifies an agent when calling the manifest rendering API
type AgentIdentifier interface {
	Value() string
}

// AgentID identifies an agent with its ID.
type AgentID string

func (id AgentID) Value() string {
	return string(id)
}

// AgentName identifies an agent with its name.
type AgentName string

func (name AgentName) Value() string {
	return string(name)
}

// AgentInfo contains all data to describe an agent that has been
// registered in Bcloud.
type AgentInfo struct {
	AgentName  string `json:"agent_name"`
	AgentID    string `json:"agent_id"`
	IsNewAgent bool   `json:"is_new_agent"`
}

type Client interface {
	RegisterAgent(ctx context.Context, agentName string) (*AgentInfo, error)
	Credentials(ctx context.Context, agentID string) credentials.PerRPCCredentials
	GetAgentManifest(ctx context.Context, identifier AgentIdentifier) (string, error)
}

type client struct {
	clientID        string
	clientSecret    string
	tokenAuthConfig *clientcredentials.Config
	base            url.URL
	noTLS           bool
}

// New creates a new ApiClient
func New(clientID, clientSecret, apiAddr string, noTLS bool) Client {
	addrScheme := "https"
	if noTLS {
		addrScheme = "http"
	}

	base := url.URL{Scheme: addrScheme, Host: apiAddr}

	tokenURL := base
	tokenURL.Path = tokenEndpoint

	tokenAuthConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL.String(),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return &client{
		clientID:        clientID,
		clientSecret:    clientSecret,
		tokenAuthConfig: tokenAuthConfig,
		base:            base,
		noTLS:           noTLS,
	}
}

// RegisterAgent registers the agent with Bcloud
func (c *client) RegisterAgent(ctx context.Context, agentName string) (*AgentInfo, error) {
	client := c.tokenAuthConfig.Client(ctx)

	registerURL := c.base
	registerURL.Path = registerAgentEndpoint
	registerURL.RawQuery = url.Values{k8s.AgentNameKey: []string{agentName}}.Encode()

	req, err := http.NewRequest(http.MethodPut, registerURL.String(), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		genericErr := fmt.Errorf("agent registration api returned: %d", rsp.StatusCode)
		// we will try and parse the json error object here
		data, err := io.ReadAll(rsp.Body)
		if err != nil {
			return nil, genericErr
		}
		jsonErr := &jsonError{}
		if err := json.Unmarshal(data, jsonErr); err != nil {
			return nil, genericErr
		}

		return nil, errors.New(jsonErr.Error)
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

// GetAgentManifest renders a manifest for an agent
func (c *client) GetAgentManifest(ctx context.Context, identifier AgentIdentifier) (string, error) {
	client := c.tokenAuthConfig.Client(ctx)

	manifestURL := c.base
	manifestURL.Path = agentManifestEndpoint

	var queryKey string
	switch identifier.(type) {
	case AgentID:
		queryKey = k8s.AgentIDKey
	case AgentName:
		queryKey = k8s.AgentNameKey
	default:
		return "", fmt.Errorf("invalid agent identifier %+v", identifier)
	}

	manifestURL.RawQuery = url.Values{queryKey: []string{identifier.Value()}}.Encode()

	req, err := http.NewRequest(http.MethodGet, manifestURL.String(), nil)
	if err != nil {
		return "", err
	}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent manifest rendering api returned: %d", rsp.StatusCode)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Credentials returns a token source for a particular agent
func (c *client) Credentials(ctx context.Context, agentID string) credentials.PerRPCCredentials {
	agentTokenURL := c.base
	agentTokenURL.Path = agentTokenEndpoint
	agentTokenURL.RawQuery = url.Values{k8s.AgentIDKey: []string{agentID}}.Encode()

	authConfig := &clientcredentials.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		TokenURL:     agentTokenURL.String(),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	ts := authConfig.TokenSource(ctx)

	return newTokenPerRPCCreds(ts, c.noTLS)
}
