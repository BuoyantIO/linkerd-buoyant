package registrator

import (
	"context"
	"fmt"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	pkgk8s "github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/k8s"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Registrator is used to ensure that the agent that will be running on the
// cluster is fully registered. If the agent is not it will perform an automatic
// agent registration using the provided client id and client secret
// credentials.
type Registrator struct {
	bcloudApiClient        bcloudapi.Client
	k8sAPI                 *k8s.KubernetesAPI
	agentMetadataConfigMap string
}

// New creates a new Registrator.
func New(bcloudApiClient bcloudapi.Client, k8sAPI *k8s.KubernetesAPI, agentMetadataConfigMap string) *Registrator {
	return &Registrator{
		bcloudApiClient:        bcloudApiClient,
		k8sAPI:                 k8sAPI,
		agentMetadataConfigMap: agentMetadataConfigMap,
	}
}

// EnsureRegistered performes agent registration if necessary. It inspects the
// config map containing the agent metadata and decides whether this is a new agent
// that needs to be registered.
func (ar *Registrator) EnsureRegistered(ctx context.Context) (*bcloudapi.AgentInfo, error) {
	cm, err := ar.k8sAPI.
		CoreV1().
		ConfigMaps(pkgk8s.AgentNamespace).
		Get(ctx, ar.agentMetadataConfigMap, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, fmt.Errorf("could not find %s config map", ar.agentMetadataConfigMap)
		}
		return nil, err
	}

	agentID, hasAgentID := cm.Data[pkgk8s.AgentIDKey]
	agentName, hasAgentName := cm.Data[pkgk8s.AgentNameKey]

	if !hasAgentName {
		return nil, fmt.Errorf("%s config map needs to have an %s key", ar.agentMetadataConfigMap, pkgk8s.AgentNameKey)
	}

	if hasAgentID {
		log.Debugf("Agent with ID %s already registered", agentID)
		ai := &bcloudapi.AgentInfo{
			AgentName:  agentName,
			AgentID:    agentID,
			IsNewAgent: false,
		}

		return ai, nil
	}

	info, err := ar.bcloudApiClient.RegisterAgent(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("agent registration failed: %w", err)
	}

	cm.Data[pkgk8s.AgentIDKey] = info.AgentID
	_, err = ar.k8sAPI.
		CoreV1().
		ConfigMaps(pkgk8s.AgentNamespace).
		Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update %s config map: %w", ar.agentMetadataConfigMap, err)
	}

	return info, nil
}
