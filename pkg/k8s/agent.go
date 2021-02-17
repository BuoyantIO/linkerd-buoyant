package k8s

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	// Namespace is the namespace where the Buoyant Cloud agent is installed.
	Namespace = "buoyant-cloud"
	// AgentName is the name of the Buoyant Cloud Agent deployment.
	AgentName = "buoyant-cloud-agent"
	// MetricsName is the name of the Buoyant Cloud Metrics deployment.
	MetricsName = "buoyant-cloud-metrics"
	// VersionLabel is the label key for the agent's version
	VersionLabel = "app.kubernetes.io/version"

	agentSecret = "buoyant-cloud-id"
)

// Agent represents the linkerd-buoyant agent. Any of these fields may not be
// present, depending on which resources are already on the cluster.
type Agent struct {
	Name    string
	Version string
	URL     string
}

// GetAgent retrieves the linkerd-buoyant agent from Kubernetes, and returns the
// agent's name, version, and url.
func (c *client) Agent(ctx context.Context) (*Agent, error) {
	var name, version, url string

	secret, err := c.Secret(ctx)
	if err == nil {
		name = string(secret.Data["name"])
		url = fmt.Sprintf(
			"%s/agent/buoyant-cloud-k8s-%s-%s-%s.yml",
			c.bcloudServer, secret.Data["name"], secret.Data["id"], secret.Data["downloadKey"],
		)
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	} else {
		// secret not found, we can't identify the agent
		return nil, nil
	}

	deploy, err := c.Deployment(ctx, AgentName)
	if err == nil {
		version = Version(deploy)
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	}

	return &Agent{
		name,
		version,
		url,
	}, nil
}
