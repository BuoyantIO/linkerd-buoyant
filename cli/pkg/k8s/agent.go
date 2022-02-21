package k8s

import (
	"context"

	agentk8s "github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	// AgentName is the name of the Buoyant Cloud Agent deployment.
	AgentName = "buoyant-cloud-agent"
	// MetricsName is the name of the Buoyant Cloud Metrics deployment.
	MetricsName = "buoyant-cloud-metrics"
	// VersionLabel is the label key for the agent's version
	VersionLabel = "app.kubernetes.io/version"

	agentMetadataMap = "agent-metadata"
)

// Agent represents the Buoyant Cloud agent. Any of these fields may not be
// present, depending on which resources are already on the cluster.
type Agent struct {
	Name    string
	Id      string
	Version string
}

func (c *client) Agent(ctx context.Context) (*Agent, error) {
	var name, id, version string

	cm, err := c.ConfigMap(ctx)
	if err == nil {
		name = cm.Data[agentk8s.AgentNameKey]
		id = cm.Data[agentk8s.AgentIDKey]
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	} else {
		// config map not found, we can't identify the agent
		return nil, nil
	}

	deploy, err := c.Deployment(ctx, AgentName)
	if err == nil {
		version = Version(deploy)
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	}

	return &Agent{
		Name:    name,
		Id:      id,
		Version: version,
	}, nil
}
