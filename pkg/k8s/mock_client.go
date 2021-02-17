package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type MockClient struct {
	MockNamespace          *v1.Namespace
	MockClusterRole        *rbacv1.ClusterRole
	MockClusterRoleBinding *rbacv1.ClusterRoleBinding
	MockSecret             *v1.Secret
	MockServiceAccount     *v1.ServiceAccount
	MockDeployment         *appsv1.Deployment
	MockPods               *v1.PodList

	MockAgent     *Agent
	MockResources []string
}

func (m *MockClient) Namespace(ctx context.Context) (*v1.Namespace, error) {
	return m.MockNamespace, nil
}
func (m *MockClient) ClusterRole(ctx context.Context) (*rbacv1.ClusterRole, error) {
	return m.MockClusterRole, nil
}
func (m *MockClient) ClusterRoleBinding(ctx context.Context) (*rbacv1.ClusterRoleBinding, error) {
	return m.MockClusterRoleBinding, nil
}
func (m *MockClient) Secret(ctx context.Context) (*v1.Secret, error) {
	return m.MockSecret, nil
}
func (m *MockClient) ServiceAccount(ctx context.Context) (*v1.ServiceAccount, error) {
	return m.MockServiceAccount, nil
}
func (m *MockClient) Deployment(ctx context.Context, name string) (*appsv1.Deployment, error) {
	return m.MockDeployment, nil
}
func (m *MockClient) Pods(ctx context.Context, labelSelector string) (*v1.PodList, error) {
	return m.MockPods, nil
}

func (m *MockClient) Agent(ctx context.Context) (*Agent, error) {
	return m.MockAgent, nil
}
func (m *MockClient) Resources(ctx context.Context) ([]string, error) {
	return m.MockResources, nil
}
