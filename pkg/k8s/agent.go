package k8s

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	agentNamespace    = "buoyant-cloud"
	agentName         = "buoyant-cloud-agent"
	agentSecret       = "buoyant-cloud-id"
	versionAnnotation = "buoyant.cloud/version"
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
func GetAgent(ctx context.Context, client kubernetes.Interface, bcloudServer string) (*Agent, error) {
	var name, version, url string

	secret, err := client.
		CoreV1().
		Secrets(agentNamespace).
		Get(ctx, agentSecret, metav1.GetOptions{})
	if err == nil {
		name = string(secret.Data["name"])
		url = fmt.Sprintf(
			"%s/agent/buoyant-cloud-k8s-%s-%s-%s.yml",
			bcloudServer, secret.Data["name"], secret.Data["id"], secret.Data["downloadKey"],
		)
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	} else {
		// secret not found, we can't identify the agent
		return nil, nil
	}

	deploy, err := client.
		AppsV1().
		Deployments(agentNamespace).
		Get(ctx, agentName, metav1.GetOptions{})
	if err == nil {
		version = deploy.GetAnnotations()[versionAnnotation]
	} else if !kerrors.IsNotFound(err) {
		return nil, err
	}

	return &Agent{
		name,
		version,
		url,
	}, nil
}
