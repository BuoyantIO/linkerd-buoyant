package registrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/linkerd/linkerd2/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace              = "buoyant-cloud"
	agentMetadataConfigMap = "agent-metadata"
	agentIDKey             = "agent_id"
	agentNameKey           = "agent_name"
)

type AgentInfo struct {
	AgentName  string `json:"agent_name"`
	AgentID    string `json:"agent_id"`
	IsNewAgent bool   `json:"is_new_agent"`
}

// AgentRegistrator is used to ensure that the agent that will be
// running on the cluster is fully registered. If the agent is not
// it will perform an automatic agent registration using the provided
// client id and client secret credentials
type AgentRegistrator struct {
	clientID     string
	clientSecret string
	apiAddr      string
	secure       bool
	k8sAPI       *k8s.KubernetesAPI
}

func NewAgentRegistrator(clientID, clientSecret, apiAddr string, secure bool, k8sAPI *k8s.KubernetesAPI) *AgentRegistrator {
	return &AgentRegistrator{
		clientID:     clientID,
		clientSecret: clientSecret,
		apiAddr:      apiAddr,
		secure:       secure,
		k8sAPI:       k8sAPI,
	}
}

func (ar *AgentRegistrator) EnsureRegistered(ctx context.Context) (*AgentInfo, error) {
	cm, err := ar.k8sAPI.
		CoreV1().
		ConfigMaps(namespace).
		Get(ctx, agentMetadataConfigMap, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, fmt.Errorf("could not find %s config map", agentMetadataConfigMap)
		}
		return nil, err
	}

	agentID, hasAgentID := cm.Data[agentIDKey]
	agentName, hasAgentName := cm.Data[agentNameKey]

	if !hasAgentName {
		return nil, fmt.Errorf("%s config map needs to have an %s key", agentMetadataConfigMap, agentNameKey)
	}

	if hasAgentID {
		log.Debugf("Agent with ID %s already registered", agentID)
		return &AgentInfo{
			AgentName:  agentName,
			AgentID:    agentID,
			IsNewAgent: false,
		}, nil
	}

	info, err := ar.registerAgent(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("agent registration failed: %w", err)
	}

	cm.Data[agentIDKey] = info.AgentID
	_, err = ar.k8sAPI.
		CoreV1().
		ConfigMaps(namespace).
		Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update %s configm map: %w", agentMetadataConfigMap, err)
	}

	return info, nil
}

func (ar *AgentRegistrator) registerAgent(ctx context.Context, agentName string) (*AgentInfo, error) {
	addrScheme := "http"
	if ar.secure {
		addrScheme = "https"
	}

	authConfig := &clientcredentials.Config{
		ClientID:     ar.clientID,
		ClientSecret: ar.clientSecret,
		TokenURL:     fmt.Sprintf("%s://%s/token", addrScheme, ar.apiAddr),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	client := authConfig.Client(ctx)
	url := fmt.Sprintf("%s://%s/register-agent?%s=%s", addrScheme, ar.apiAddr, agentNameKey, agentName)
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
