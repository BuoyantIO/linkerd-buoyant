package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// MockClient provides a mock Kubernetes client for testing
type MockClient struct {
	MockNamespace          *v1.Namespace
	MockClusterRole        *rbacv1.ClusterRole
	MockClusterRoleBinding *rbacv1.ClusterRoleBinding
	MockConfigMap          *v1.ConfigMap
	MockServiceAccount     *v1.ServiceAccount
	MockDaemonSet          *appsv1.DaemonSet
	MockDeployment         *appsv1.Deployment
	MockPods               *v1.PodList

	MockAgent     *Agent
	MockResources []string
}

// Namespace returns a mock Namespace object.
func (m *MockClient) Namespace(ctx context.Context) (*v1.Namespace, error) {
	return m.MockNamespace, nil
}

// ClusterRole returns a mock ClusterRole object.
func (m *MockClient) ClusterRole(ctx context.Context) (*rbacv1.ClusterRole, error) {
	return m.MockClusterRole, nil
}

// ClusterRoleBinding returns a mock ClusterRoleBinding object.
func (m *MockClient) ClusterRoleBinding(ctx context.Context) (*rbacv1.ClusterRoleBinding, error) {
	return m.MockClusterRoleBinding, nil
}

// Secret returns a mock ConfigMap object.
func (m *MockClient) ConfigMap(ctx context.Context) (*v1.ConfigMap, error) {
	return m.MockConfigMap, nil
}

// ServiceAccount returns a mock ServiceAccount object.
func (m *MockClient) ServiceAccount(ctx context.Context) (*v1.ServiceAccount, error) {
	return m.MockServiceAccount, nil
}

// DaemonSet returns a mock DaemonSet object.
func (m *MockClient) DaemonSet(ctx context.Context, name string) (*appsv1.DaemonSet, error) {
	return m.MockDaemonSet, nil
}

// Deployment returns a mock Deployment object.
func (m *MockClient) Deployment(ctx context.Context, name string) (*appsv1.Deployment, error) {
	return m.MockDeployment, nil
}

// Pods returns a mock Pod List.
func (m *MockClient) Pods(ctx context.Context, labelSelector string) (*v1.PodList, error) {
	return m.MockPods, nil
}

// Agent returns a mock Buoyant Cloud agent.
func (m *MockClient) Agent(ctx context.Context) (*Agent, error) {
	return m.MockAgent, nil
}

// Resources returns mock Buoyant Cloud agent resources.
func (m *MockClient) Resources(ctx context.Context) ([]string, error) {
	return m.MockResources, nil
}
