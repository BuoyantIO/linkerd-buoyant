package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	// Client defines the interface for linkerd-buoyant's Kubernetes client
	Client interface {
		// Namespace retrieves the buoyant-cloud Namespace.
		Namespace(ctx context.Context) (*v1.Namespace, error)
		// ClusterRole retrieves the buoyant-cloud-agent CR.
		ClusterRole(ctx context.Context) (*rbacv1.ClusterRole, error)
		// ClusterRoleBinding retrieves the buoyant-cloud-agent CRB.
		ClusterRoleBinding(ctx context.Context) (*rbacv1.ClusterRoleBinding, error)
		// Secret retrieves the buoyant-cloud-id Secret.
		Secret(ctx context.Context) (*v1.Secret, error)
		// ServiceAccount retrieves the buoyant-cloud-agent ServiceAccount.
		ServiceAccount(ctx context.Context) (*v1.ServiceAccount, error)
		// Deployment retrieves a Deployment by name in the buoyant-cloud namespace.
		Deployment(ctx context.Context, name string) (*appsv1.Deployment, error)
		// Pods retrieves a PodList by labelSelector from the buoyant-cloud
		// namespace.
		Pods(ctx context.Context, labelSelector string) (*v1.PodList, error)

		// Agent retrieves the Buoyant Cloud agent from Kubernetes, and returns the
		// agent's name, version, and url. If Agent is not found, it will return a
		// nil Agent with no error.
		Agent(ctx context.Context) (*Agent, error)
		// Resources returns all linkerd-buoyant resources required for deletion.
		// Specifically, these three resource types matching the label
		// app.kubernetes.io/part-of=buoyant-cloud:
		// - ClusterRole
		// - ClusterRoleBinding
		// - Namespace
		Resources(ctx context.Context) ([]string, error)
	}

	// client is the internal struct satisfying the Client interface
	client struct {
		kubernetes.Interface
		bcloudServer string
	}
)

// New takes a kubeconfig and kubecontext and returns an initialized Client.
func New(kubeconfig, kubecontext, bcloudServer string) (Client, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: kubecontext})

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &client{clientset, bcloudServer}, nil
}

func (c *client) Namespace(ctx context.Context) (*v1.Namespace, error) {
	return c.
		CoreV1().
		Namespaces().
		Get(ctx, Namespace, metav1.GetOptions{})
}

func (c *client) ClusterRole(ctx context.Context) (*rbacv1.ClusterRole, error) {
	return c.
		RbacV1().
		ClusterRoles().
		Get(ctx, AgentName, metav1.GetOptions{})
}

func (c *client) ClusterRoleBinding(ctx context.Context) (*rbacv1.ClusterRoleBinding, error) {
	return c.
		RbacV1().
		ClusterRoleBindings().
		Get(ctx, AgentName, metav1.GetOptions{})
}

func (c *client) Secret(ctx context.Context) (*v1.Secret, error) {
	return c.
		CoreV1().
		Secrets(Namespace).
		Get(ctx, agentSecret, metav1.GetOptions{})
}

func (c *client) ServiceAccount(ctx context.Context) (*v1.ServiceAccount, error) {
	return c.
		CoreV1().
		ServiceAccounts(Namespace).
		Get(ctx, AgentName, metav1.GetOptions{})
}

func (c *client) Deployment(ctx context.Context, name string) (*appsv1.Deployment, error) {
	return c.
		AppsV1().
		Deployments(Namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func (c *client) Pods(ctx context.Context, labelSelector string) (*v1.PodList, error) {
	return c.
		CoreV1().
		Pods(Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
}

// Version returns the label value of app.kubernetes.io/version.
func Version(deploy *appsv1.Deployment) string {
	if deploy == nil {
		return ""
	}
	return deploy.GetLabels()[VersionLabel]
}
